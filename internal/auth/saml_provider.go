package auth

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/crewjam/saml"
	"github.com/crewjam/saml/samlsp"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"

	"github.com/nimbleflux/fluxbase/internal/database"
)

func (s *SAMLService) RemoveProvider(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.providers, name)
	delete(s.spConfigs, name)

	log.Info().Str("provider", name).Msg("SAML provider removed")
}

func (s *SAMLService) LoadProvidersFromDB(ctx context.Context) error {
	query := `
		SELECT id, name, enabled, entity_id, acs_url,
		       idp_metadata_url, idp_metadata_xml, idp_metadata_cached,
		       attribute_mapping, auto_create_users, default_role,
		       COALESCE(allow_dashboard_login, false), COALESCE(allow_app_login, true),
		       COALESCE(allow_idp_initiated, false), COALESCE(allowed_redirect_hosts, ARRAY[]::TEXT[]),
		       created_at, updated_at
		FROM auth.saml_providers
		WHERE enabled = true AND COALESCE(source, 'database') = 'database'
	`

	if pingErr := s.db.Pool().Ping(ctx); pingErr != nil {
		return fmt.Errorf("failed to ping database before loading SAML providers: %w", pingErr)
	}

	return database.WrapWithServiceRole(ctx, s.db, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx, query)
		if err != nil {
			return fmt.Errorf("failed to query SAML providers: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var (
				id                   string
				name                 string
				enabled              bool
				entityID             string
				acsURL               string
				metadataURL          *string
				metadataXML          *string
				metadataCached       *string
				attrMapping          map[string]string
				autoCreateUsers      bool
				defaultRole          string
				allowDashboardLogin  bool
				allowAppLogin        bool
				allowIDPInitiated    bool
				allowedRedirectHosts []string
				createdAt            time.Time
				updatedAt            time.Time
			)

			err := rows.Scan(
				&id, &name, &enabled, &entityID, &acsURL,
				&metadataURL, &metadataXML, &metadataCached,
				&attrMapping, &autoCreateUsers, &defaultRole,
				&allowDashboardLogin, &allowAppLogin,
				&allowIDPInitiated, &allowedRedirectHosts,
				&createdAt, &updatedAt,
			)
			if err != nil {
				log.Error().Err(err).Msg("Failed to scan SAML provider from database")
				continue
			}

			s.mu.RLock()
			_, exists := s.providers[name]
			s.mu.RUnlock()
			if exists {
				log.Debug().Str("provider", name).Msg("Skipping DB provider - already loaded from config")
				continue
			}

			var metadataToUse string
			//nolint:gocritic // Conditions check different variables, not switch-compatible
			if metadataCached != nil && *metadataCached != "" {
				metadataToUse = *metadataCached
			} else if metadataXML != nil && *metadataXML != "" {
				metadataToUse = *metadataXML
			} else if metadataURL != nil && *metadataURL != "" {
				xmlData, err := s.fetchMetadata(*metadataURL)
				if err != nil {
					log.Warn().Err(err).Str("provider", name).Msg("Failed to fetch SAML metadata from URL")
					continue
				}
				metadataToUse = string(xmlData)
			} else {
				log.Warn().Str("provider", name).Msg("No SAML metadata available")
				continue
			}

			metadata, err := samlsp.ParseMetadata([]byte(metadataToUse))
			if err != nil {
				log.Warn().Err(err).Str("provider", name).Msg("Failed to parse SAML metadata")
				continue
			}

			var idpDescriptor *saml.IDPSSODescriptor
			for i := range metadata.IDPSSODescriptors {
				desc := &metadata.IDPSSODescriptors[i]
				for _, sso := range desc.SingleSignOnServices {
					if sso.Binding == saml.HTTPPostBinding || sso.Binding == saml.HTTPRedirectBinding {
						idpDescriptor = desc
						break
					}
				}
				if idpDescriptor != nil {
					break
				}
			}
			if idpDescriptor == nil {
				log.Warn().Str("provider", name).Msg("No suitable IdP SSO descriptor found")
				continue
			}

			var ssoURL string
			for _, sso := range idpDescriptor.SingleSignOnServices {
				if sso.Binding == saml.HTTPPostBinding || sso.Binding == saml.HTTPRedirectBinding {
					ssoURL = sso.Location
					break
				}
			}

			var sloURL string
			for _, slo := range idpDescriptor.SingleLogoutServices {
				if slo.Binding == saml.HTTPPostBinding || slo.Binding == saml.HTTPRedirectBinding {
					sloURL = slo.Location
					break
				}
			}

			var certificate string
			for _, kd := range idpDescriptor.KeyDescriptors {
				if kd.Use == "signing" || kd.Use == "" {
					for _, cert := range kd.KeyInfo.X509Data.X509Certificates {
						certificate = cert.Data
						break
					}
					break
				}
			}

			provider := &SAMLProvider{
				ID:                     id,
				Name:                   name,
				Enabled:                enabled,
				EntityID:               entityID,
				AcsURL:                 acsURL,
				SsoURL:                 ssoURL,
				SloURL:                 sloURL,
				Certificate:            certificate,
				AttributeMapping:       attrMapping,
				AutoCreateUsers:        autoCreateUsers,
				DefaultRole:            defaultRole,
				AllowIDPInitiated:      allowIDPInitiated,
				AllowedRedirectHosts:   allowedRedirectHosts,
				CreatedAt:              createdAt,
				UpdatedAt:              updatedAt,
				idpDescriptor:          idpDescriptor,
				metadata:               metadata,
				AllowDashboardLogin:    allowDashboardLogin,
				AllowAppLogin:          allowAppLogin,
				RequireLogoutSignature: true,
			}

			acsURLParsed, _ := url.Parse(acsURL)
			entityIDParsed, _ := url.Parse(entityID)
			metadataURLParsed, _ := url.Parse(fmt.Sprintf("%s/auth/saml/metadata/%s", s.baseURL, name))

			sp := &saml.ServiceProvider{
				EntityID:          entityIDParsed.String(),
				AcsURL:            *acsURLParsed,
				MetadataURL:       *metadataURLParsed,
				IDPMetadata:       metadata,
				AllowIDPInitiated: allowIDPInitiated,
			}

			s.mu.Lock()
			s.providers[name] = provider
			s.spConfigs[name] = sp
			s.mu.Unlock()

			log.Info().Str("provider", name).Msg("Loaded SAML provider from database")
		}

		return nil
	})
}

func (s *SAMLService) ReloadProviderFromDB(ctx context.Context, name string) error {
	query := `
		SELECT id, name, enabled, entity_id, acs_url,
		       idp_metadata_url, idp_metadata_xml, idp_metadata_cached,
		       attribute_mapping, auto_create_users, default_role,
		       COALESCE(allow_dashboard_login, false), COALESCE(allow_app_login, true),
		       COALESCE(allow_idp_initiated, false), COALESCE(allowed_redirect_hosts, ARRAY[]::TEXT[]),
		       created_at, updated_at
		FROM auth.saml_providers
		WHERE name = $1
	`

	return database.WrapWithServiceRole(ctx, s.db, func(tx pgx.Tx) error {
		var (
			id                   string
			providerName         string
			enabled              bool
			entityID             string
			acsURL               string
			metadataURL          *string
			metadataXML          *string
			metadataCached       *string
			attrMapping          map[string]string
			autoCreateUsers      bool
			defaultRole          string
			allowDashboardLogin  bool
			allowAppLogin        bool
			allowIDPInitiated    bool
			allowedRedirectHosts []string
			createdAt            time.Time
			updatedAt            time.Time
		)

		err := tx.QueryRow(ctx, query, name).Scan(
			&id, &providerName, &enabled, &entityID, &acsURL,
			&metadataURL, &metadataXML, &metadataCached,
			&attrMapping, &autoCreateUsers, &defaultRole,
			&allowDashboardLogin, &allowAppLogin,
			&allowIDPInitiated, &allowedRedirectHosts,
			&createdAt, &updatedAt,
		)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				s.RemoveProvider(name)
				return nil
			}
			return fmt.Errorf("failed to query SAML provider: %w", err)
		}

		if !enabled {
			s.RemoveProvider(name)
			return nil
		}

		provider := &SAMLProvider{
			ID:                   id,
			Name:                 providerName,
			Enabled:              enabled,
			EntityID:             entityID,
			AcsURL:               acsURL,
			AttributeMapping:     attrMapping,
			AutoCreateUsers:      autoCreateUsers,
			DefaultRole:          defaultRole,
			AllowDashboardLogin:  allowDashboardLogin,
			AllowAppLogin:        allowAppLogin,
			AllowIDPInitiated:    allowIDPInitiated,
			AllowedRedirectHosts: allowedRedirectHosts,
			CreatedAt:            createdAt,
			UpdatedAt:            updatedAt,
		}

		if provider.DefaultRole == "" {
			provider.DefaultRole = "authenticated"
		}
		provider.RequireLogoutSignature = true

		if err := s.loadProviderMetadata(provider, metadataCached, metadataXML, metadataURL); err != nil {
			return fmt.Errorf("failed to load SAML metadata for provider %s: %w", name, err)
		}

		acsURLParsed, _ := url.Parse(acsURL)
		entityIDParsed, _ := url.Parse(entityID)
		metadataURLParsed, _ := url.Parse(fmt.Sprintf("%s/auth/saml/metadata/%s", s.baseURL, providerName))

		sp := &saml.ServiceProvider{
			EntityID:          entityIDParsed.String(),
			AcsURL:            *acsURLParsed,
			MetadataURL:       *metadataURLParsed,
			IDPMetadata:       provider.metadata,
			AllowIDPInitiated: allowIDPInitiated,
		}

		s.mu.Lock()
		s.providers[providerName] = provider
		s.spConfigs[providerName] = sp
		s.mu.Unlock()

		log.Info().Str("provider", providerName).Msg("Reloaded SAML provider from database")
		return nil
	})
}

func (s *SAMLService) GetProvidersForDashboard() []*SAMLProvider {
	s.mu.RLock()
	defer s.mu.RUnlock()

	providers := make([]*SAMLProvider, 0)
	for _, p := range s.providers {
		if p.Enabled && p.AllowDashboardLogin {
			providers = append(providers, p)
		}
	}

	return providers
}

func (s *SAMLService) GetProvidersForApp() []*SAMLProvider {
	s.mu.RLock()
	defer s.mu.RUnlock()

	providers := make([]*SAMLProvider, 0)
	for _, p := range s.providers {
		if p.Enabled && p.AllowAppLogin {
			providers = append(providers, p)
		}
	}

	return providers
}

func (s *SAMLService) GetProviderForTenant(ctx context.Context, name string, tenantID string) (*SAMLProvider, error) {
	if tenantID != "" && s.db != nil {
		var provider SAMLProvider
		var metadataCached *string
		var metadataURL *string
		var metadataXML *string
		var attrMapping map[string]string
		var allowedRedirectHosts []string

		query := `
			SELECT id, name, enabled, entity_id, acs_url,
				   idp_metadata_url, idp_metadata_xml, idp_metadata_cached,
				   attribute_mapping, auto_create_users, default_role,
				   COALESCE(allow_dashboard_login, false), COALESCE(allow_app_login, true),
				   COALESCE(allow_idp_initiated, false), COALESCE(allowed_redirect_hosts, ARRAY[]::TEXT[]),
				   created_at, updated_at
			FROM auth.saml_providers
			WHERE name = $1 AND tenant_id = $2::uuid AND enabled = true
		`
		err := database.WrapWithServiceRole(ctx, s.db, func(tx pgx.Tx) error {
			return tx.QueryRow(ctx, query, name, tenantID).Scan(
				&provider.ID, &provider.Name, &provider.Enabled, &provider.EntityID, &provider.AcsURL,
				&metadataURL, &metadataXML, &metadataCached,
				&attrMapping, &provider.AutoCreateUsers, &provider.DefaultRole,
				&provider.AllowDashboardLogin, &provider.AllowAppLogin,
				&provider.AllowIDPInitiated, &allowedRedirectHosts,
				&provider.CreatedAt, &provider.UpdatedAt,
			)
		})
		if err == nil {
			provider.AttributeMapping = attrMapping
			provider.AllowedRedirectHosts = allowedRedirectHosts
			provider.GroupAttribute = "groups"
			provider.RequireLogoutSignature = true

			if err := s.loadProviderMetadata(&provider, metadataCached, metadataXML, metadataURL); err != nil {
				log.Warn().Err(err).Str("provider", name).Msg("Failed to load tenant SAML metadata")
			}

			return &provider, nil
		}
	}

	return s.GetProvider(name)
}

func (s *SAMLService) GetProvidersForAppWithTenant(ctx context.Context, tenantID string) []*SAMLProvider {
	providers := make([]*SAMLProvider, 0)

	if tenantID != "" && s.db != nil {
		query := `
			SELECT id, name, enabled, entity_id, acs_url,
				   COALESCE(allow_app_login, true) as allow_app_login
			FROM auth.saml_providers
			WHERE tenant_id = $1::uuid AND enabled = true AND COALESCE(allow_app_login, true) = true
		`
		_ = database.WrapWithServiceRole(ctx, s.db, func(tx pgx.Tx) error {
			rows, err := tx.Query(ctx, query, tenantID)
			if err != nil {
				return err
			}
			defer rows.Close()
			for rows.Next() {
				var p SAMLProvider
				if err := rows.Scan(&p.ID, &p.Name, &p.Enabled, &p.EntityID, &p.AcsURL, &p.AllowAppLogin); err == nil {
					providers = append(providers, &p)
				}
			}
			return nil
		})
	}

	s.mu.RLock()
	for _, p := range s.providers {
		if p.Enabled && p.AllowAppLogin {
			found := false
			for _, tp := range providers {
				if tp.Name == p.Name {
					found = true
					break
				}
			}
			if !found {
				providers = append(providers, p)
			}
		}
	}
	s.mu.RUnlock()

	return providers
}

func (s *SAMLService) GetProvidersForDashboardWithTenant(ctx context.Context, tenantID string) []*SAMLProvider {
	providers := make([]*SAMLProvider, 0)

	if tenantID != "" && s.db != nil {
		query := `
			SELECT id, name, enabled, entity_id, acs_url,
				   COALESCE(allow_dashboard_login, false) as allow_dashboard_login
			FROM auth.saml_providers
			WHERE tenant_id = $1::uuid AND enabled = true AND COALESCE(allow_dashboard_login, false) = true
		`
		_ = database.WrapWithServiceRole(ctx, s.db, func(tx pgx.Tx) error {
			rows, err := tx.Query(ctx, query, tenantID)
			if err != nil {
				return err
			}
			defer rows.Close()
			for rows.Next() {
				var p SAMLProvider
				if err := rows.Scan(&p.ID, &p.Name, &p.Enabled, &p.EntityID, &p.AcsURL, &p.AllowDashboardLogin); err == nil {
					providers = append(providers, &p)
				}
			}
			return nil
		})
	}

	s.mu.RLock()
	for _, p := range s.providers {
		if p.Enabled && p.AllowDashboardLogin {
			found := false
			for _, tp := range providers {
				if tp.Name == p.Name {
					found = true
					break
				}
			}
			if !found {
				providers = append(providers, p)
			}
		}
	}
	s.mu.RUnlock()

	return providers
}

func (s *SAMLService) loadProviderMetadata(provider *SAMLProvider, metadataCached *string, metadataXML *string, metadataURL *string) error {
	var metadataToUse string

	if metadataCached != nil && *metadataCached != "" {
		metadataToUse = *metadataCached
	} else if metadataXML != nil && *metadataXML != "" {
		metadataToUse = *metadataXML
	} else if metadataURL != nil && *metadataURL != "" {
		xmlData, err := s.fetchMetadata(*metadataURL)
		if err != nil {
			return err
		}
		metadataToUse = string(xmlData)
	} else {
		return errors.New("no metadata available")
	}

	metadata, err := samlsp.ParseMetadata([]byte(metadataToUse))
	if err != nil {
		return err
	}

	for i := range metadata.IDPSSODescriptors {
		desc := &metadata.IDPSSODescriptors[i]
		for _, sso := range desc.SingleSignOnServices {
			if sso.Binding == saml.HTTPPostBinding || sso.Binding == saml.HTTPRedirectBinding {
				provider.idpDescriptor = desc
				provider.metadata = metadata

				for _, sso := range desc.SingleSignOnServices {
					if sso.Binding == saml.HTTPPostBinding || sso.Binding == saml.HTTPRedirectBinding {
						provider.SsoURL = sso.Location
						break
					}
				}

				for _, slo := range desc.SingleLogoutServices {
					if slo.Binding == saml.HTTPPostBinding || slo.Binding == saml.HTTPRedirectBinding {
						provider.IdPSloURL = slo.Location
						break
					}
				}

				for _, kd := range desc.KeyDescriptors {
					if kd.Use == "signing" || kd.Use == "" {
						for _, cert := range kd.KeyInfo.X509Data.X509Certificates {
							provider.Certificate = cert.Data
							break
						}
						break
					}
				}

				return nil
			}
		}
	}

	return errors.New("no suitable IdP SSO descriptor found")
}
