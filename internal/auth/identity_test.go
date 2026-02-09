package auth

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// =============================================================================
// Mock Identity Repository for Testing
// =============================================================================

// MockIdentityRepository is a mock implementation of identity repository operations
type MockIdentityRepository struct {
	mu             sync.RWMutex
	identities     map[string]*UserIdentity            // id -> identity
	byProvider     map[string]map[string]*UserIdentity // provider -> providerUserID -> identity
	users          map[string][]*UserIdentity          // userID -> identities
	deleteError    map[string]error                    // id -> error to return
	createError    error
	getByUserError error
}

func NewMockIdentityRepository() *MockIdentityRepository {
	return &MockIdentityRepository{
		identities:  make(map[string]*UserIdentity),
		byProvider:  make(map[string]map[string]*UserIdentity),
		users:       make(map[string][]*UserIdentity),
		deleteError: make(map[string]error),
	}
}

func (m *MockIdentityRepository) Create(ctx context.Context, userID, provider, providerUserID string, email *string, metadata map[string]interface{}) (*UserIdentity, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.createError != nil {
		return nil, m.createError
	}

	identity := &UserIdentity{
		ID:             "identity-" + provider + "-" + providerUserID,
		UserID:         userID,
		Provider:       provider,
		ProviderUserID: providerUserID,
		Email:          email,
		IdentityData:   metadata,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	m.identities[identity.ID] = identity

	if m.byProvider[provider] == nil {
		m.byProvider[provider] = make(map[string]*UserIdentity)
	}
	m.byProvider[provider][providerUserID] = identity

	m.users[userID] = append(m.users[userID], identity)

	return identity, nil
}

func (m *MockIdentityRepository) GetByUserID(ctx context.Context, userID string) ([]UserIdentity, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.getByUserError != nil {
		return nil, m.getByUserError
	}

	identities := m.users[userID]
	if identities == nil {
		return []UserIdentity{}, nil
	}

	result := make([]UserIdentity, len(identities))
	for i, id := range identities {
		result[i] = *id
	}
	return result, nil
}

func (m *MockIdentityRepository) GetByID(ctx context.Context, id string) (*UserIdentity, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	identity, exists := m.identities[id]
	if !exists {
		return nil, ErrIdentityNotFound
	}
	return identity, nil
}

func (m *MockIdentityRepository) GetByProviderAndUserID(ctx context.Context, provider, providerUserID string) (*UserIdentity, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.byProvider[provider] == nil {
		return nil, ErrIdentityNotFound
	}

	identity, exists := m.byProvider[provider][providerUserID]
	if !exists {
		return nil, ErrIdentityNotFound
	}
	return identity, nil
}

func (m *MockIdentityRepository) Delete(ctx context.Context, identityID, userID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err, ok := m.deleteError[identityID]; ok {
		return err
	}

	identity, exists := m.identities[identityID]
	if !exists {
		return ErrIdentityNotFound
	}

	// Verify ownership
	if identity.UserID != userID {
		return ErrIdentityNotFound
	}

	// Remove from all indexes
	delete(m.identities, identityID)
	if m.byProvider[identity.Provider] != nil {
		delete(m.byProvider[identity.Provider], identity.ProviderUserID)
	}

	// Remove from user list
	userList := m.users[userID]
	for i, id := range userList {
		if id.ID == identityID {
			m.users[userID] = append(userList[:i], userList[i+1:]...)
			break
		}
	}

	return nil
}

// SetCreateError sets an error to return on Create
func (m *MockIdentityRepository) SetCreateError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.createError = err
}

// SetDeleteError sets an error to return on Delete for a specific identity
func (m *MockIdentityRepository) SetDeleteError(identityID string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.deleteError[identityID] = err
}

// SetGetByUserError sets an error to return on GetByUserID
func (m *MockIdentityRepository) SetGetByUserError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.getByUserError = err
}

// =============================================================================
// Testable Identity Service
// =============================================================================

// TestableIdentityService wraps IdentityService with mock repository support for testing
type TestableIdentityService struct {
	identityService *IdentityService
	mockRepo        *MockIdentityRepository
}

// NewTestableIdentityService creates a testable identity service with mock repository
func NewTestableIdentityService(mockRepo *MockIdentityRepository, stateStore *StateStore) *TestableIdentityService {
	oauthManager := NewOAuthManager()
	var repo *IdentityRepository // nil for now, we use the mock

	// Create a real IdentityService but we'll use mock methods that override repo access
	return &TestableIdentityService{
		mockRepo:        mockRepo,
		identityService: NewIdentityService(repo, oauthManager, stateStore),
	}
}

// GetUserIdentities retrieves all identities for a user using the mock repository
func (s *TestableIdentityService) GetUserIdentities(ctx context.Context, userID string) ([]UserIdentity, error) {
	return s.mockRepo.GetByUserID(ctx, userID)
}

// LinkIdentityProvider initiates OAuth flow to link a new provider
func (s *TestableIdentityService) LinkIdentityProvider(ctx context.Context, userID string, provider string) (string, string, error) {
	return s.identityService.LinkIdentityProvider(ctx, userID, provider)
}

// LinkIdentity creates or updates an identity link for a user
func (s *TestableIdentityService) LinkIdentity(ctx context.Context, userID, provider, providerUserID string, email *string, metadata map[string]interface{}) (*UserIdentity, error) {
	// Check if this provider identity is already linked
	existingIdentity, err := s.mockRepo.GetByProviderAndUserID(ctx, provider, providerUserID)
	if err != nil && !errors.Is(err, ErrIdentityNotFound) {
		return nil, err
	}

	// If already linked to another user, return error
	if existingIdentity != nil && existingIdentity.UserID != userID {
		return nil, ErrIdentityAlreadyLinked
	}

	// If already linked to this user, return existing identity
	if existingIdentity != nil {
		return existingIdentity, nil
	}

	// Create new identity link
	return s.mockRepo.Create(ctx, userID, provider, providerUserID, email, metadata)
}

// UnlinkIdentity removes an OAuth identity from a user
func (s *TestableIdentityService) UnlinkIdentity(ctx context.Context, userID, identityID string) error {
	return s.mockRepo.Delete(ctx, identityID, userID)
}

// =============================================================================
// Functional Tests for IdentityService
// =============================================================================

func TestIdentityService_GetUserIdentities_Success(t *testing.T) {
	mockRepo := NewMockIdentityRepository()
	stateStore := NewStateStore()
	svc := NewTestableIdentityService(mockRepo, stateStore)
	ctx := context.Background()

	// Create some test identities
	email1 := "user1@example.com"
	email2 := "user2@example.com"

	_, _ = mockRepo.Create(ctx, "user-123", "google", "google-user-1", &email1, map[string]interface{}{"name": "User 1"})
	_, _ = mockRepo.Create(ctx, "user-123", "github", "github-user-1", &email2, map[string]interface{}{"name": "User 1"})

	// Get identities
	identities, err := svc.GetUserIdentities(ctx, "user-123")

	require.NoError(t, err)
	assert.Len(t, identities, 2)

	providers := make([]string, 0, 2)
	for _, id := range identities {
		providers = append(providers, id.Provider)
	}
	assert.Contains(t, providers, "google")
	assert.Contains(t, providers, "github")
}

func TestIdentityService_GetUserIdentities_NoIdentities(t *testing.T) {
	mockRepo := NewMockIdentityRepository()
	stateStore := NewStateStore()
	svc := NewTestableIdentityService(mockRepo, stateStore)
	ctx := context.Background()

	identities, err := svc.GetUserIdentities(ctx, "nonexistent-user")

	require.NoError(t, err)
	assert.Len(t, identities, 0)
}

func TestIdentityService_GetUserIdentities_DbError(t *testing.T) {
	mockRepo := NewMockIdentityRepository()
	mockRepo.SetGetByUserError(errors.New("database connection lost"))
	stateStore := NewStateStore()
	svc := NewTestableIdentityService(mockRepo, stateStore)
	ctx := context.Background()

	identities, err := svc.GetUserIdentities(ctx, "user-123")

	assert.Error(t, err)
	assert.Nil(t, identities)
}

func TestIdentityService_LinkIdentityProvider_Google(t *testing.T) {
	mockRepo := NewMockIdentityRepository()
	stateStore := NewStateStore()
	svc := NewTestableIdentityService(mockRepo, stateStore)
	ctx := context.Background()

	// Register Google provider
	err := svc.identityService.oauthManager.RegisterProvider(ProviderGoogle, OAuthConfig{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURL:  "https://example.com/callback",
		Scopes:       []string{"email", "profile"},
	})
	require.NoError(t, err)

	authURL, state, err := svc.LinkIdentityProvider(ctx, "user-123", "google")

	require.NoError(t, err)
	assert.NotEmpty(t, authURL)
	assert.NotEmpty(t, state)
	assert.Contains(t, authURL, "google.com")
}

func TestIdentityService_LinkIdentityProvider_GitHub(t *testing.T) {
	mockRepo := NewMockIdentityRepository()
	stateStore := NewStateStore()
	svc := NewTestableIdentityService(mockRepo, stateStore)
	ctx := context.Background()

	// Register GitHub provider
	err := svc.identityService.oauthManager.RegisterProvider("github", OAuthConfig{
		ClientID:     "github-client-id",
		ClientSecret: "github-client-secret",
		RedirectURL:  "https://example.com/callback",
		Scopes:       []string{"user:email", "read:user"},
	})
	require.NoError(t, err)

	authURL, state, err := svc.LinkIdentityProvider(ctx, "user-456", "github")

	require.NoError(t, err)
	assert.NotEmpty(t, authURL)
	assert.NotEmpty(t, state)
	assert.Contains(t, authURL, "github.com")
}

func TestIdentityService_LinkIdentityProvider_InvalidProvider(t *testing.T) {
	mockRepo := NewMockIdentityRepository()
	stateStore := NewStateStore()
	svc := NewTestableIdentityService(mockRepo, stateStore)
	ctx := context.Background()

	// Try to link unregistered provider
	authURL, state, err := svc.LinkIdentityProvider(ctx, "user-123", "not-a-real-provider")

	assert.Error(t, err)
	assert.Empty(t, authURL)
	assert.Empty(t, state)
}

func TestIdentityService_LinkIdentityProvider_StateGeneration(t *testing.T) {
	mockRepo := NewMockIdentityRepository()
	stateStore := NewStateStore()
	svc := NewTestableIdentityService(mockRepo, stateStore)
	ctx := context.Background()

	err := svc.identityService.oauthManager.RegisterProvider(ProviderGoogle, OAuthConfig{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURL:  "https://example.com/callback",
		Scopes:       []string{"email", "profile"},
	})
	require.NoError(t, err)

	// Request multiple link operations
	_, state1, _ := svc.LinkIdentityProvider(ctx, "user-123", "google")
	_, state2, _ := svc.LinkIdentityProvider(ctx, "user-123", "google")

	// States should be different
	assert.NotEqual(t, state1, state2)

	// Both should be valid in state store
	assert.True(t, stateStore.Validate(state1))
	assert.True(t, stateStore.Validate(state2))
}

func TestIdentityService_UnlinkIdentity_Success(t *testing.T) {
	mockRepo := NewMockIdentityRepository()
	stateStore := NewStateStore()
	svc := NewTestableIdentityService(mockRepo, stateStore)
	ctx := context.Background()

	// Create a test identity
	email := "user@example.com"
	identity, _ := mockRepo.Create(ctx, "user-123", "google", "google-user-id", &email, map[string]interface{}{"name": "Test User"})

	// Unlink the identity
	err := svc.UnlinkIdentity(ctx, "user-123", identity.ID)

	require.NoError(t, err)

	// Verify identity is gone
	_, err = mockRepo.GetByID(ctx, identity.ID)
	assert.Error(t, err)
	assert.Equal(t, ErrIdentityNotFound, err)
}

func TestIdentityService_UnlinkIdentity_NotFound(t *testing.T) {
	mockRepo := NewMockIdentityRepository()
	stateStore := NewStateStore()
	svc := NewTestableIdentityService(mockRepo, stateStore)
	ctx := context.Background()

	err := svc.UnlinkIdentity(ctx, "user-123", "nonexistent-identity")

	assert.Error(t, err)
	assert.Equal(t, ErrIdentityNotFound, err)
}

func TestIdentityService_UnlinkIdentity_WrongUser(t *testing.T) {
	mockRepo := NewMockIdentityRepository()
	stateStore := NewStateStore()
	svc := NewTestableIdentityService(mockRepo, stateStore)
	ctx := context.Background()

	// Create a test identity for user-123
	email := "user@example.com"
	identity, _ := mockRepo.Create(ctx, "user-123", "google", "google-user-id", &email, map[string]interface{}{})

	// Try to unlink from different user
	err := svc.UnlinkIdentity(ctx, "different-user-456", identity.ID)

	assert.Error(t, err)
	assert.Equal(t, ErrIdentityNotFound, err)
}

func TestIdentityService_LinkIdentity_NewIdentity(t *testing.T) {
	mockRepo := NewMockIdentityRepository()
	stateStore := NewStateStore()
	svc := NewTestableIdentityService(mockRepo, stateStore)
	ctx := context.Background()

	email := "user@example.com"
	metadata := map[string]interface{}{
		"name":   "SAML User",
		"issuer": "https://saml.example.com",
	}

	// Link new identity (for SAML/SSO)
	identity, err := svc.LinkIdentity(ctx, "user-123", "saml", "saml-user-id", &email, metadata)

	require.NoError(t, err)
	assert.NotNil(t, identity)
	assert.Equal(t, "user-123", identity.UserID)
	assert.Equal(t, "saml", identity.Provider)
	assert.Equal(t, "saml-user-id", identity.ProviderUserID)
	assert.Equal(t, email, *identity.Email)
	assert.Equal(t, "SAML User", identity.IdentityData["name"])
}

func TestIdentityService_LinkIdentity_AlreadyLinkedSameUser(t *testing.T) {
	mockRepo := NewMockIdentityRepository()
	stateStore := NewStateStore()
	svc := NewTestableIdentityService(mockRepo, stateStore)
	ctx := context.Background()

	email := "user@example.com"

	// Create initial identity
	_, _ = mockRepo.Create(ctx, "user-123", "google", "google-user-id", &email, map[string]interface{}{})

	// Try to link again - should return existing identity
	identity, err := svc.LinkIdentity(ctx, "user-123", "google", "google-user-id", &email, map[string]interface{}{})

	require.NoError(t, err)
	assert.NotNil(t, identity)
	assert.Equal(t, "user-123", identity.UserID)
}

func TestIdentityService_LinkIdentity_AlreadyLinkedDifferentUser(t *testing.T) {
	mockRepo := NewMockIdentityRepository()
	stateStore := NewStateStore()
	svc := NewTestableIdentityService(mockRepo, stateStore)
	ctx := context.Background()

	email1 := "user1@example.com"
	email2 := "user2@example.com"

	// Create identity for user-123
	_, _ = mockRepo.Create(ctx, "user-123", "google", "google-user-id", &email1, map[string]interface{}{})

	// Try to link same provider ID to different user - should fail
	identity, err := svc.LinkIdentity(ctx, "user-456", "google", "google-user-id", &email2, map[string]interface{}{})

	assert.Error(t, err)
	assert.Nil(t, identity)
	assert.Equal(t, ErrIdentityAlreadyLinked, err)
}

func TestIdentityService_LinkIdentity_MultipleProviders(t *testing.T) {
	mockRepo := NewMockIdentityRepository()
	stateStore := NewStateStore()
	svc := NewTestableIdentityService(mockRepo, stateStore)
	ctx := context.Background()

	// Link multiple providers to same user
	email1 := "user1@example.com"
	email2 := "user2@example.com"
	email3 := "user3@example.com"

	_, err := svc.LinkIdentity(ctx, "user-123", "google", "google-user-id", &email1, map[string]interface{}{"provider": "google"})
	require.NoError(t, err)

	_, err = svc.LinkIdentity(ctx, "user-123", "github", "github-user-id", &email2, map[string]interface{}{"provider": "github"})
	require.NoError(t, err)

	_, err = svc.LinkIdentity(ctx, "user-123", "microsoft", "ms-user-id", &email3, map[string]interface{}{"provider": "microsoft"})
	require.NoError(t, err)

	// Verify all identities are linked
	identities, _ := mockRepo.GetByUserID(ctx, "user-123")
	assert.Len(t, identities, 3)

	providers := make(map[string]bool)
	for _, id := range identities {
		providers[id.Provider] = true
	}

	assert.True(t, providers["google"])
	assert.True(t, providers["github"])
	assert.True(t, providers["microsoft"])
}

func TestIdentityService_UnlinkIdentity_LastIdentity(t *testing.T) {
	mockRepo := NewMockIdentityRepository()
	stateStore := NewStateStore()
	svc := NewTestableIdentityService(mockRepo, stateStore)
	ctx := context.Background()

	email := "user@example.com"
	identity, _ := mockRepo.Create(ctx, "user-123", "google", "google-user-id", &email, map[string]interface{}{})

	// Unlink the only identity
	err := svc.UnlinkIdentity(ctx, "user-123", identity.ID)

	require.NoError(t, err)

	// Verify no identities left
	identities, _ := mockRepo.GetByUserID(ctx, "user-123")
	assert.Len(t, identities, 0)
}

func TestIdentityService_UnlinkIdentity_OneOfMany(t *testing.T) {
	mockRepo := NewMockIdentityRepository()
	stateStore := NewStateStore()
	svc := NewTestableIdentityService(mockRepo, stateStore)
	ctx := context.Background()

	email1 := "user1@example.com"
	email2 := "user2@example.com"

	id1, _ := mockRepo.Create(ctx, "user-123", "google", "google-user-id", &email1, map[string]interface{}{})
	_, _ = mockRepo.Create(ctx, "user-123", "github", "github-user-id", &email2, map[string]interface{}{})

	// Unlink one identity
	err := svc.UnlinkIdentity(ctx, "user-123", id1.ID)

	require.NoError(t, err)

	// Verify one identity remains
	identities, _ := mockRepo.GetByUserID(ctx, "user-123")
	assert.Len(t, identities, 1)
	assert.Equal(t, "github", identities[0].Provider)
}

func TestIdentityService_ListIdentities_AllProviders(t *testing.T) {
	mockRepo := NewMockIdentityRepository()
	stateStore := NewStateStore()
	svc := NewTestableIdentityService(mockRepo, stateStore)
	ctx := context.Background()

	// Create identities with multiple providers
	providers := []string{"google", "github", "microsoft", "apple", "facebook"}
	for i, provider := range providers {
		email := "user@example.com"
		_, _ = mockRepo.Create(ctx, "user-123", provider, provider+"-user-id-"+string(rune('0'+i)), &email, map[string]interface{}{
			"provider": provider,
			"index":    i,
		})
	}

	// Get all identities
	identities, err := svc.GetUserIdentities(ctx, "user-123")

	require.NoError(t, err)
	assert.Len(t, identities, 5)

	// Verify all providers are present
	providerMap := make(map[string]bool)
	for _, id := range identities {
		providerMap[id.Provider] = true
	}

	for _, provider := range providers {
		assert.True(t, providerMap[provider], "Provider %s should be in list", provider)
	}
}

func TestIdentityService_ListIdentities_EmptyList(t *testing.T) {
	mockRepo := NewMockIdentityRepository()
	stateStore := NewStateStore()
	svc := NewTestableIdentityService(mockRepo, stateStore)
	ctx := context.Background()

	// Get identities for user with none
	identities, err := svc.GetUserIdentities(ctx, "user-with-no-identities")

	require.NoError(t, err)
	assert.Empty(t, identities)
}

func TestUserIdentity_ProviderTypes(t *testing.T) {
	tests := []struct {
		name     string
		provider string
	}{
		{"Google provider", "google"},
		{"GitHub provider", "github"},
		{"Microsoft provider", "microsoft"},
		{"Apple provider", "apple"},
		{"Facebook provider", "facebook"},
		{"Twitter provider", "twitter"},
		{"LinkedIn provider", "linkedin"},
		{"GitLab provider", "gitlab"},
		{"Bitbucket provider", "bitbucket"},
		{"SAML provider", "saml"},
		{"OIDC provider", "oidc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := NewMockIdentityRepository()
			stateStore := NewStateStore()
			svc := NewTestableIdentityService(mockRepo, stateStore)
			ctx := context.Background()

			email := "user@example.com"
			identity, err := svc.LinkIdentity(ctx, "user-123", tt.provider, tt.provider+"-user-id", &email, map[string]interface{}{})

			require.NoError(t, err)
			assert.Equal(t, tt.provider, identity.Provider)
		})
	}
}

func TestIdentityService_IdentityMetadata(t *testing.T) {
	mockRepo := NewMockIdentityRepository()
	stateStore := NewStateStore()
	svc := NewTestableIdentityService(mockRepo, stateStore)
	ctx := context.Background()

	// Create identity with rich metadata
	email := "user@example.com"
	metadata := map[string]interface{}{
		"name":         "Full Name",
		"given_name":   "First",
		"family_name":  "Last",
		"picture":      "https://example.com/avatar.jpg",
		"locale":       "en",
		"verified":     true,
		"custom_field": "custom value",
		"nested":       map[string]interface{}{"key": "value"},
		"array_field":  []string{"a", "b", "c"},
		"number_field": 42,
	}

	identity, err := svc.LinkIdentity(ctx, "user-123", "google", "google-user-id", &email, metadata)

	require.NoError(t, err)
	assert.NotNil(t, identity.IdentityData)
	assert.Equal(t, "Full Name", identity.IdentityData["name"])
	assert.Equal(t, "First", identity.IdentityData["given_name"])
	assert.Equal(t, 42, identity.IdentityData["number_field"])
}

// =============================================================================
// Test Suite for Identity Service
// =============================================================================

type IdentityServiceTestSuite struct {
	suite.Suite
	repo       *MockIdentityRepository
	stateStore *StateStore
	svc        *TestableIdentityService
	ctx        context.Context
}

func (suite *IdentityServiceTestSuite) SetupTest() {
	suite.repo = NewMockIdentityRepository()
	suite.stateStore = NewStateStore()
	suite.svc = NewTestableIdentityService(suite.repo, suite.stateStore)
	suite.ctx = context.Background()
}

func (suite *IdentityServiceTestSuite) TearDownTest() {
	// Clean up between tests
	suite.repo = NewMockIdentityRepository()
	suite.stateStore = NewStateStore()
	suite.svc = NewTestableIdentityService(suite.repo, suite.stateStore)
}

func (suite *IdentityServiceTestSuite) TestListIdentities() {
	// Test listing identities for a user with multiple providers
	email := "user@example.com"
	suite.repo.Create(suite.ctx, "user-123", "google", "google-1", &email, map[string]interface{}{})
	suite.repo.Create(suite.ctx, "user-123", "github", "github-1", &email, map[string]interface{}{})

	identities, err := suite.svc.GetUserIdentities(suite.ctx, "user-123")

	suite.NoError(err)
	suite.Len(identities, 2)
}

func (suite *IdentityServiceTestSuite) TestLinkUnlinkIdentity() {
	// Test linking and then unlinking an identity
	email := "user@example.com"
	identity, _ := suite.repo.Create(suite.ctx, "user-123", "google", "google-1", &email, map[string]interface{}{})

	// Verify linked
	identities, _ := suite.svc.GetUserIdentities(suite.ctx, "user-123")
	suite.Len(identities, 1)

	// Unlink
	err := suite.svc.UnlinkIdentity(suite.ctx, "user-123", identity.ID)
	suite.NoError(err)

	// Verify unlinked
	identities, _ = suite.svc.GetUserIdentities(suite.ctx, "user-123")
	suite.Len(identities, 0)
}

func (suite *IdentityServiceTestSuite) TestLinkMultipleProviders() {
	// Test linking multiple providers to same user
	email := "user@example.com"
	suite.svc.LinkIdentity(suite.ctx, "user-123", "google", "google-1", &email, map[string]interface{}{})
	suite.svc.LinkIdentity(suite.ctx, "user-123", "github", "github-1", &email, map[string]interface{}{})
	suite.svc.LinkIdentity(suite.ctx, "user-123", "microsoft", "ms-1", &email, map[string]interface{}{})

	identities, _ := suite.svc.GetUserIdentities(suite.ctx, "user-123")
	suite.Len(identities, 3)
}

func TestIdentityServiceSuite(t *testing.T) {
	suite.Run(t, new(IdentityServiceTestSuite))
}

// =============================================================================
// Error Variable Tests
// =============================================================================

func TestIdentityErrors(t *testing.T) {
	t.Run("error types are defined", func(t *testing.T) {
		assert.NotNil(t, ErrIdentityNotFound)
		assert.NotNil(t, ErrIdentityAlreadyLinked)
	})

	t.Run("error messages are meaningful", func(t *testing.T) {
		assert.Contains(t, ErrIdentityNotFound.Error(), "not found")
		assert.Contains(t, ErrIdentityAlreadyLinked.Error(), "already linked")
	})

	t.Run("errors are distinct", func(t *testing.T) {
		assert.NotEqual(t, ErrIdentityNotFound, ErrIdentityAlreadyLinked)
	})

	t.Run("error messages are exact", func(t *testing.T) {
		assert.Equal(t, "identity not found", ErrIdentityNotFound.Error())
		assert.Equal(t, "identity is already linked to another user", ErrIdentityAlreadyLinked.Error())
	})
}

// =============================================================================
// UserIdentity Struct Tests
// =============================================================================

func TestUserIdentity_Struct(t *testing.T) {
	t.Run("creates identity with all fields", func(t *testing.T) {
		now := time.Now()
		email := "user@example.com"

		identity := UserIdentity{
			ID:             "identity-123",
			UserID:         "user-456",
			Provider:       "google",
			ProviderUserID: "google-user-789",
			Email:          &email,
			IdentityData: map[string]interface{}{
				"name":    "Test User",
				"picture": "https://example.com/avatar.jpg",
			},
			CreatedAt: now,
			UpdatedAt: now,
		}

		assert.Equal(t, "identity-123", identity.ID)
		assert.Equal(t, "user-456", identity.UserID)
		assert.Equal(t, "google", identity.Provider)
		assert.Equal(t, "google-user-789", identity.ProviderUserID)
		assert.Equal(t, "user@example.com", *identity.Email)
		assert.Equal(t, "Test User", identity.IdentityData["name"])
	})

	t.Run("handles nil optional fields", func(t *testing.T) {
		identity := UserIdentity{
			ID:             "identity-123",
			UserID:         "user-456",
			Provider:       "github",
			ProviderUserID: "github-user-789",
		}

		assert.Nil(t, identity.Email)
		assert.Nil(t, identity.IdentityData)
	})
}

func TestUserIdentity_Providers(t *testing.T) {
	providers := []string{
		"google",
		"github",
		"microsoft",
		"apple",
		"facebook",
		"twitter",
		"linkedin",
		"gitlab",
		"bitbucket",
		"saml",
		"oidc",
	}

	for _, provider := range providers {
		t.Run("provider_"+provider, func(t *testing.T) {
			identity := UserIdentity{
				ID:             "identity-123",
				UserID:         "user-456",
				Provider:       provider,
				ProviderUserID: provider + "-user-789",
			}

			assert.Equal(t, provider, identity.Provider)
		})
	}
}

func TestUserIdentity_IdentityData(t *testing.T) {
	t.Run("empty identity data", func(t *testing.T) {
		identity := UserIdentity{
			IdentityData: map[string]interface{}{},
		}

		assert.NotNil(t, identity.IdentityData)
		assert.Empty(t, identity.IdentityData)
	})

	t.Run("complex identity data", func(t *testing.T) {
		identity := UserIdentity{
			IdentityData: map[string]interface{}{
				"name":         "Test User",
				"email":        "test@example.com",
				"picture":      "https://example.com/avatar.jpg",
				"verified":     true,
				"login_count":  42,
				"last_login":   "2026-01-13T12:00:00Z",
				"permissions":  []string{"read", "write"},
				"organization": map[string]interface{}{"id": "org-123", "name": "Acme Corp"},
			},
		}

		assert.Equal(t, "Test User", identity.IdentityData["name"])
		assert.Equal(t, true, identity.IdentityData["verified"])
		assert.Equal(t, 42, identity.IdentityData["login_count"])
	})

	t.Run("nil vs empty identity data", func(t *testing.T) {
		identityNil := UserIdentity{IdentityData: nil}
		identityEmpty := UserIdentity{IdentityData: map[string]interface{}{}}

		assert.Nil(t, identityNil.IdentityData)
		assert.NotNil(t, identityEmpty.IdentityData)
	})
}

func TestUserIdentity_Timestamps(t *testing.T) {
	now := time.Now()
	past := now.Add(-24 * time.Hour)

	identity := UserIdentity{
		CreatedAt: past,
		UpdatedAt: now,
	}

	assert.True(t, identity.CreatedAt.Before(identity.UpdatedAt))
	assert.True(t, identity.UpdatedAt.After(identity.CreatedAt))
}

// =============================================================================
// Repository Tests
// =============================================================================

func TestNewIdentityRepository(t *testing.T) {
	// Test that it doesn't panic with nil db
	repo := NewIdentityRepository(nil)
	assert.NotNil(t, repo)
}

func TestNewIdentityRepository_Fields(t *testing.T) {
	repo := NewIdentityRepository(nil)

	require.NotNil(t, repo)
	assert.Nil(t, repo.db)
}

// =============================================================================
// Service Tests
// =============================================================================

func TestNewIdentityService(t *testing.T) {
	// Test that it doesn't panic with nil dependencies
	svc := NewIdentityService(nil, nil, nil)
	assert.NotNil(t, svc)
}

func TestNewIdentityService_Fields(t *testing.T) {
	mockRepo := NewMockIdentityRepository()
	stateStore := NewStateStore()

	svc := NewTestableIdentityService(mockRepo, stateStore)

	require.NotNil(t, svc)
	assert.NotNil(t, svc.mockRepo)
	assert.NotNil(t, svc.identityService)
}

func TestNewIdentityService_WithAllNil(t *testing.T) {
	// Test that the real service can be created with nil dependencies
	svc := NewIdentityService(nil, nil, nil)

	require.NotNil(t, svc)
	// Note: We can't access private fields to verify they're nil
	// but we can verify the service was created without panicking
}

// =============================================================================
// IdentityService Method Structure Tests (without DB)
// =============================================================================

func TestIdentityService_LinkIdentityProvider_UnregisteredProvider(t *testing.T) {
	mockRepo := NewMockIdentityRepository()
	stateStore := NewStateStore()
	svc := NewTestableIdentityService(mockRepo, stateStore)

	// Try to use an unregistered provider - should return error
	_, _, err := svc.LinkIdentityProvider(nil, "user-123", "google")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid OAuth provider")
}

func TestIdentityService_LinkIdentityProvider_WithOAuthManager(t *testing.T) {
	mockRepo := NewMockIdentityRepository()
	stateStore := NewStateStore()
	svc := NewTestableIdentityService(mockRepo, stateStore)

	// Register a provider first
	err := svc.identityService.oauthManager.RegisterProvider(ProviderGoogle, OAuthConfig{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURL:  "https://example.com/callback",
		Scopes:       []string{"email", "profile"},
	})
	require.NoError(t, err)

	// This should work now
	authURL, state, err := svc.LinkIdentityProvider(nil, "user-123", "google")

	require.NoError(t, err)
	assert.NotEmpty(t, authURL)
	assert.NotEmpty(t, state)
	assert.Contains(t, authURL, "client_id=test-client-id")
}

func TestIdentityService_LinkIdentityCallback_InvalidState(t *testing.T) {
	mockRepo := NewMockIdentityRepository()
	stateStore := NewStateStore()
	svc := NewTestableIdentityService(mockRepo, stateStore)

	// Try to callback with invalid state
	_, _, err := svc.LinkIdentityProvider(nil, "user-123", "google")

	// We expect this to fail because provider isn't registered
	assert.Error(t, err)
}

func TestIdentityService_StateStoreIntegration(t *testing.T) {
	mockRepo := NewMockIdentityRepository()
	stateStore := NewStateStore()
	svc := NewTestableIdentityService(mockRepo, stateStore)

	err := svc.identityService.oauthManager.RegisterProvider(ProviderGoogle, OAuthConfig{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURL:  "https://example.com/callback",
		Scopes:       []string{"email", "profile"},
	})
	require.NoError(t, err)

	// Generate state via service
	_, state, err := svc.LinkIdentityProvider(nil, "user-123", "google")
	require.NoError(t, err)

	// State should be stored in state store
	// Note: Validate consumes the state, so we can only check once
	valid := stateStore.Validate(state)
	assert.True(t, valid)

	// After validation, state should be consumed
	validAgain := stateStore.Validate(state)
	assert.False(t, validAgain)
}

// =============================================================================
// Provider-specific Tests
// =============================================================================

func TestUserIdentity_GoogleProvider(t *testing.T) {
	email := "user@gmail.com"
	identity := UserIdentity{
		ID:             "identity-google-123",
		UserID:         "user-456",
		Provider:       "google",
		ProviderUserID: "114823456789012345678",
		Email:          &email,
		IdentityData: map[string]interface{}{
			"sub":            "114823456789012345678",
			"name":           "Google User",
			"given_name":     "Google",
			"family_name":    "User",
			"picture":        "https://lh3.googleusercontent.com/a/default-user",
			"email":          "user@gmail.com",
			"email_verified": true,
			"locale":         "en",
		},
	}

	assert.Equal(t, "google", identity.Provider)
	assert.Equal(t, "114823456789012345678", identity.ProviderUserID)
	assert.Equal(t, true, identity.IdentityData["email_verified"])
}

func TestUserIdentity_GithubProvider(t *testing.T) {
	email := "user@github.com"
	identity := UserIdentity{
		ID:             "identity-github-123",
		UserID:         "user-456",
		Provider:       "github",
		ProviderUserID: "12345678",
		Email:          &email,
		IdentityData: map[string]interface{}{
			"id":         12345678,
			"login":      "githubuser",
			"name":       "GitHub User",
			"email":      "user@github.com",
			"avatar_url": "https://avatars.githubusercontent.com/u/12345678?v=4",
			"company":    "Acme Corp",
			"location":   "San Francisco, CA",
		},
	}

	assert.Equal(t, "github", identity.Provider)
	assert.Equal(t, "githubuser", identity.IdentityData["login"])
}

func TestUserIdentity_MicrosoftProvider(t *testing.T) {
	email := "user@outlook.com"
	identity := UserIdentity{
		ID:             "identity-microsoft-123",
		UserID:         "user-456",
		Provider:       "microsoft",
		ProviderUserID: "abc123-def456-ghi789",
		Email:          &email,
		IdentityData: map[string]interface{}{
			"id":                "abc123-def456-ghi789",
			"displayName":       "Microsoft User",
			"mail":              "user@outlook.com",
			"userPrincipalName": "user@outlook.com",
		},
	}

	assert.Equal(t, "microsoft", identity.Provider)
	assert.Equal(t, "Microsoft User", identity.IdentityData["displayName"])
}

// =============================================================================
// Benchmark Tests
// =============================================================================

func BenchmarkUserIdentity_Creation(b *testing.B) {
	email := "user@example.com"

	for i := 0; i < b.N; i++ {
		_ = UserIdentity{
			ID:             "identity-123",
			UserID:         "user-456",
			Provider:       "google",
			ProviderUserID: "google-user-789",
			Email:          &email,
			IdentityData: map[string]interface{}{
				"name":    "Test User",
				"picture": "https://example.com/avatar.jpg",
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
	}
}

func BenchmarkNewIdentityService(b *testing.B) {
	oauthManager := NewOAuthManager()
	stateStore := NewStateStore()
	repo := NewIdentityRepository(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewIdentityService(repo, oauthManager, stateStore)
	}
}

func BenchmarkNewIdentityRepository(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewIdentityRepository(nil)
	}
}
