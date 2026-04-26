package functions

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"

	"github.com/nimbleflux/fluxbase/internal/database"
	"github.com/nimbleflux/fluxbase/internal/middleware"
)

// RegisterAdminRoutes registers admin-only routes for functions management
// These routes should be called with UnifiedAuthMiddleware and RequireRole("admin", "instance_admin")
func (h *Handler) RegisterAdminRoutes(app *fiber.App) {
	// Admin-only function reload endpoint
	app.Post("/api/v1/admin/functions/reload", h.ReloadFunctions)
}

// bundleFunctionFromFilesystem loads function code with supporting files and shared modules,
// then bundles it. Returns bundled code, original code, bundled status, and any error.
func (h *Handler) bundleFunctionFromFilesystem(ctx context.Context, functionName string) (bundledCode string, originalCode string, isBundled bool, bundleError *string, err error) {
	// Load main code and supporting files
	mainCode, supportingFiles, err := LoadFunctionCodeWithFiles(h.functionsDir, functionName)
	if err != nil {
		return "", "", false, nil, fmt.Errorf("failed to load code: %w", err)
	}

	// Load shared modules from filesystem
	sharedModules, err := LoadSharedModulesFromFilesystem(h.functionsDir)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to load shared modules from filesystem, continuing without them")
		sharedModules = make(map[string]string)
	}

	// Create bundler
	bundler, bundlerErr := h.createBundler()
	if bundlerErr != nil {
		// No bundler available - return unbundled code
		return mainCode, mainCode, false, nil, nil
	}

	// Determine if we need to use BundleWithFiles (multi-file or shared imports)
	hasSharedImports := strings.Contains(mainCode, "from \"_shared/") || strings.Contains(mainCode, "from '_shared/")
	hasMultipleFiles := len(supportingFiles) > 0

	var result *BundleResult
	var bundleErr error

	if hasSharedImports || hasMultipleFiles {
		// Use BundleWithFiles for multi-file or shared module support
		result, bundleErr = bundler.BundleWithFiles(ctx, mainCode, supportingFiles, sharedModules)
	} else {
		// Simple single-file bundle
		result, bundleErr = bundler.Bundle(ctx, mainCode)
	}

	if bundleErr != nil {
		// Bundling failed - return unbundled code with error
		errMsg := fmt.Sprintf("bundle failed: %v", bundleErr)
		return mainCode, mainCode, false, &errMsg, nil
	}

	// Bundling succeeded
	var bundleErrPtr *string
	if result.Error != "" {
		bundleErrPtr = &result.Error
	}

	return result.BundledCode, mainCode, result.IsBundled, bundleErrPtr, nil
}

// ReloadFunctions scans the functions directory and syncs with database
// Admin-only endpoint - requires authentication and admin role
func (h *Handler) ReloadFunctions(c fiber.Ctx) error {
	ctx := middleware.CtxWithTenant(c)

	syncCtx := database.ContextWithTenant(ctx, "")
	currentTenantID := database.TenantFromContext(ctx)

	functionFiles, err := ListFunctionFiles(h.functionsDir)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to scan functions directory",
		})
	}

	allFunctions, err := h.storage.ListFunctionsForSync(syncCtx, currentTenantID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to list existing functions",
		})
	}

	// Build set of function names on disk
	diskFunctionNames := make(map[string]bool)
	for _, fileInfo := range functionFiles {
		diskFunctionNames[fileInfo.Name] = true
	}

	// Track results
	created := []string{}
	updated := []string{}
	deleted := []string{}
	errors := []string{}

	// Process each function file
	for _, fileInfo := range functionFiles {
		// Check if function exists in database
		existingFn, err := h.storage.GetFunctionForSync(syncCtx, fileInfo.Name, currentTenantID)

		if err != nil {
			// Function doesn't exist in database - create it
			bundledCode, originalCode, isBundled, bundleError, err := h.bundleFunctionFromFilesystem(ctx, fileInfo.Name)
			if err != nil {
				errors = append(errors, fmt.Sprintf("%s: %v", fileInfo.Name, err))
				continue
			}

			// Parse configuration from code comments
			config := ParseFunctionConfig(originalCode)

			// Create new function with default settings
			fn := &EdgeFunction{
				Name:                 fileInfo.Name,
				Code:                 bundledCode,
				OriginalCode:         &originalCode,
				IsBundled:            isBundled,
				BundleError:          bundleError,
				Enabled:              true,
				TimeoutSeconds:       30,
				MemoryLimitMB:        128,
				AllowNet:             true,
				AllowEnv:             true,
				AllowRead:            false,
				AllowWrite:           false,
				AllowUnauthenticated: config.AllowUnauthenticated,
				IsPublic:             config.IsPublic,
				DisableExecutionLogs: config.DisableExecutionLogs,
				Source:               "filesystem",
			}

			if err := h.storage.CreateFunction(ctx, fn); err != nil {
				errors = append(errors, fmt.Sprintf("%s: failed to create: %v", fileInfo.Name, err))
				continue
			}

			created = append(created, fileInfo.Name)
		} else {
			// Function exists - update code from filesystem
			bundledCode, originalCode, isBundled, bundleError, err := h.bundleFunctionFromFilesystem(ctx, fileInfo.Name)
			if err != nil {
				errors = append(errors, fmt.Sprintf("%s: %v", fileInfo.Name, err))
				continue
			}

			// Parse configuration from code comments
			config := ParseFunctionConfig(originalCode)

			// Update if code or config has changed
			// Compare with original_code if available, otherwise with code
			compareCode := originalCode
			if existingFn.OriginalCode != nil {
				compareCode = *existingFn.OriginalCode
			}

			if existingFn.Code != bundledCode || compareCode != originalCode || existingFn.AllowUnauthenticated != config.AllowUnauthenticated || existingFn.IsPublic != config.IsPublic || existingFn.DisableExecutionLogs != config.DisableExecutionLogs {
				updates := map[string]interface{}{
					"code":                   bundledCode,
					"original_code":          originalCode,
					"is_bundled":             isBundled,
					"bundle_error":           bundleError,
					"allow_unauthenticated":  config.AllowUnauthenticated,
					"is_public":              config.IsPublic,
					"disable_execution_logs": config.DisableExecutionLogs,
				}

				if err := h.storage.UpdateFunctionForSync(syncCtx, fileInfo.Name, currentTenantID, updates); err != nil {
					errors = append(errors, fmt.Sprintf("%s: failed to update: %v", fileInfo.Name, err))
					continue
				}

				updated = append(updated, fileInfo.Name)
			}
		}
	}

	// Delete functions that exist in database but not on disk
	// Only delete filesystem-sourced functions, preserve API-created ones
	for _, dbFunc := range allFunctions {
		if !diskFunctionNames[dbFunc.Name] && dbFunc.Source == "filesystem" {
			// Function exists in DB but not on disk - delete it
			if err := h.storage.DeleteFunctionForSync(syncCtx, dbFunc.Name, "default", currentTenantID); err != nil {
				errors = append(errors, fmt.Sprintf("%s: failed to delete: %v", dbFunc.Name, err))
				continue
			}
			deleted = append(deleted, dbFunc.Name)
		}
	}

	return c.JSON(fiber.Map{
		"message": "Functions reloaded from filesystem",
		"created": created,
		"updated": updated,
		"deleted": deleted,
		"errors":  errors,
		"total":   len(functionFiles),
	})
}

// SyncFunctions syncs a list of functions to a specific namespace
// Admin-only endpoint - requires authentication and admin role
func (h *Handler) SyncFunctions(c fiber.Ctx) error {
	var req struct {
		Namespace string `json:"namespace"`
		Functions []struct {
			Name                 string  `json:"name"`
			Description          *string `json:"description"`
			Code                 string  `json:"code"`
			OriginalCode         *string `json:"original_code"` // Original code if pre-bundled (for editing)
			IsBundled            *bool   `json:"is_bundled"`    // If true, skip server-side bundling
			Enabled              *bool   `json:"enabled"`
			TimeoutSeconds       *int    `json:"timeout_seconds"`
			MemoryLimitMB        *int    `json:"memory_limit_mb"`
			AllowNet             *bool   `json:"allow_net"`
			AllowEnv             *bool   `json:"allow_env"`
			AllowRead            *bool   `json:"allow_read"`
			AllowWrite           *bool   `json:"allow_write"`
			AllowUnauthenticated *bool   `json:"allow_unauthenticated"`
			IsPublic             *bool   `json:"is_public"`
			CronSchedule         *string `json:"cron_schedule"`
		} `json:"functions"`
		Options struct {
			DeleteMissing bool `json:"delete_missing"`
			DryRun        bool `json:"dry_run"`
		} `json:"options"`
	}

	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Default namespace to "default" if not specified
	namespace := req.Namespace
	if namespace == "" {
		namespace = "default"
	}

	ctx := middleware.CtxWithTenant(c)

	syncCtx := database.ContextWithTenant(ctx, "")

	var createdBy *uuid.UUID
	if userID := c.Locals("user_id"); userID != nil {
		if uid, ok := userID.(string); ok {
			parsed, err := uuid.Parse(uid)
			if err == nil {
				createdBy = &parsed
			}
		}
	}

	currentTenantID := database.TenantFromContext(ctx)
	existingFunctions, err := h.storage.ListFunctionsByNamespaceForSync(syncCtx, namespace, currentTenantID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to list existing functions in namespace",
		})
	}

	// Build set of existing function names
	existingNames := make(map[string]*EdgeFunctionSummary)
	for i := range existingFunctions {
		existingNames[existingFunctions[i].Name] = &existingFunctions[i]
	}

	// Build set of payload function names
	payloadNames := make(map[string]bool)
	for _, spec := range req.Functions {
		payloadNames[spec.Name] = true
	}

	// Determine operations
	toCreate := []string{}
	toUpdate := []string{}
	toDelete := []string{}

	for _, spec := range req.Functions {
		if _, exists := existingNames[spec.Name]; exists {
			toUpdate = append(toUpdate, spec.Name)
		} else {
			toCreate = append(toCreate, spec.Name)
		}
	}

	if req.Options.DeleteMissing {
		for name := range existingNames {
			if !payloadNames[name] {
				toDelete = append(toDelete, name)
			}
		}
	}

	// Track results
	created := []string{}
	updated := []string{}
	deleted := []string{}
	unchanged := []string{}
	errorList := []fiber.Map{}

	// If dry run, return what would be done without making changes
	if req.Options.DryRun {
		return c.JSON(fiber.Map{
			"message":   "Dry run - no changes made",
			"namespace": namespace,
			"summary": fiber.Map{
				"created":   len(toCreate),
				"updated":   len(toUpdate),
				"deleted":   len(toDelete),
				"unchanged": 0,
			},
			"details": fiber.Map{
				"created":   toCreate,
				"updated":   toUpdate,
				"deleted":   toDelete,
				"unchanged": []string{},
			},
			"errors":  []string{},
			"dry_run": true,
		})
	}

	// Bundle and create/update functions in parallel
	type bundleResult struct {
		Name         string
		BundledCode  string
		OriginalCode string
		IsBundled    bool
		BundleError  *string
		Err          error
	}

	// Use semaphore to limit concurrent bundling to 10
	sem := make(chan struct{}, 10)
	var wg sync.WaitGroup
	resultsChan := make(chan bundleResult, len(req.Functions))

	// Load shared modules once (used by all bundles)
	sharedModules, _ := h.storage.ListSharedModules(ctx)
	sharedModulesMap := make(map[string]string)
	for _, module := range sharedModules {
		sharedModulesMap[module.ModulePath] = module.Content
	}

	// Bundle all functions in parallel
	for i := range req.Functions {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			sem <- struct{}{}        // Acquire semaphore
			defer func() { <-sem }() // Release semaphore

			spec := req.Functions[i]

			bundledCode := spec.Code
			originalCode := spec.Code
			isBundled := false
			var bundleError *string

			// If client sent pre-bundled code, skip server-side bundling
			if spec.IsBundled != nil && *spec.IsBundled {
				// Code is already bundled by the client
				isBundled = true
				// Use original_code if provided (for editing), otherwise use code as both
				if spec.OriginalCode != nil && *spec.OriginalCode != "" {
					originalCode = *spec.OriginalCode
				}
				resultsChan <- bundleResult{
					Name:         spec.Name,
					BundledCode:  bundledCode,
					OriginalCode: originalCode,
					IsBundled:    isBundled,
					BundleError:  bundleError,
					Err:          nil,
				}
				return
			}

			// Bundle the function code server-side
			bundler, err := h.createBundler()
			if err != nil {
				resultsChan <- bundleResult{
					Name: spec.Name,
					Err:  fmt.Errorf("failed to create bundler: %w", err),
				}
				return
			}

			// Check if code imports from _shared/ modules
			hasSharedImports := strings.Contains(spec.Code, "from \"_shared/") ||
				strings.Contains(spec.Code, "from '_shared/")

			var result *BundleResult
			var bundleErr error

			if hasSharedImports {
				supportingFiles := make(map[string]string)
				result, bundleErr = bundler.BundleWithFiles(context.Background(), spec.Code, supportingFiles, sharedModulesMap)
			} else {
				result, bundleErr = bundler.Bundle(context.Background(), spec.Code)
			}

			if bundleErr != nil {
				resultsChan <- bundleResult{
					Name: spec.Name,
					Err:  fmt.Errorf("bundle error: %w", bundleErr),
				}
				return
			}

			if result != nil {
				bundledCode = result.BundledCode
				isBundled = result.IsBundled
				if result.Error != "" {
					bundleError = &result.Error
				}
			}

			resultsChan <- bundleResult{
				Name:         spec.Name,
				BundledCode:  bundledCode,
				OriginalCode: originalCode,
				IsBundled:    isBundled,
				BundleError:  bundleError,
				Err:          nil,
			}
		}(i)
	}

	// Wait for all bundling to complete
	wg.Wait()
	close(resultsChan)

	// Collect bundling results
	bundleResults := make(map[string]bundleResult)
	for result := range resultsChan {
		bundleResults[result.Name] = result
		if result.Err != nil {
			errorList = append(errorList, fiber.Map{
				"function": result.Name,
				"error":    result.Err.Error(),
				"action":   "bundle",
			})
		}
	}

	// Create/Update functions
	for _, spec := range req.Functions {
		result, ok := bundleResults[spec.Name]
		if !ok || result.Err != nil {
			// Skip if bundling failed
			continue
		}

		// Parse configuration from code comments
		config := ParseFunctionConfig(spec.Code)

		// Determine values (request takes precedence over config)
		allowUnauthenticated := config.AllowUnauthenticated
		if spec.AllowUnauthenticated != nil {
			allowUnauthenticated = *spec.AllowUnauthenticated
		}

		isPublic := config.IsPublic
		if spec.IsPublic != nil {
			isPublic = *spec.IsPublic
		}

		if _, exists := existingNames[spec.Name]; exists {
			// Update existing function
			updates := map[string]interface{}{
				"code":                  result.BundledCode,
				"original_code":         result.OriginalCode,
				"is_bundled":            result.IsBundled,
				"bundle_error":          result.BundleError,
				"allow_unauthenticated": allowUnauthenticated,
				"is_public":             isPublic,
			}

			if spec.Description != nil {
				updates["description"] = spec.Description
			}
			if spec.Enabled != nil {
				updates["enabled"] = *spec.Enabled
			}
			if spec.TimeoutSeconds != nil {
				updates["timeout_seconds"] = *spec.TimeoutSeconds
			}
			if spec.MemoryLimitMB != nil {
				updates["memory_limit_mb"] = *spec.MemoryLimitMB
			}
			if spec.AllowNet != nil {
				updates["allow_net"] = *spec.AllowNet
			}
			if spec.AllowEnv != nil {
				updates["allow_env"] = *spec.AllowEnv
			}
			if spec.AllowRead != nil {
				updates["allow_read"] = *spec.AllowRead
			}
			if spec.AllowWrite != nil {
				updates["allow_write"] = *spec.AllowWrite
			}
			if spec.CronSchedule != nil {
				updates["cron_schedule"] = *spec.CronSchedule
			}

			if err := h.storage.UpdateFunctionByNamespaceForSync(syncCtx, spec.Name, namespace, currentTenantID, updates); err != nil {
				errorList = append(errorList, fiber.Map{
					"function": spec.Name,
					"error":    err.Error(),
					"action":   "update",
				})
				continue
			}

			updated = append(updated, spec.Name)
		} else {
			// Create new function
			fn := &EdgeFunction{
				Name:                 spec.Name,
				Namespace:            namespace,
				Description:          spec.Description,
				Code:                 result.BundledCode,
				OriginalCode:         &result.OriginalCode,
				IsBundled:            result.IsBundled,
				BundleError:          result.BundleError,
				Enabled:              valueOr(spec.Enabled, true),
				TimeoutSeconds:       valueOr(spec.TimeoutSeconds, 30),
				MemoryLimitMB:        valueOr(spec.MemoryLimitMB, 128),
				AllowNet:             valueOr(spec.AllowNet, true),
				AllowEnv:             valueOr(spec.AllowEnv, true),
				AllowRead:            valueOr(spec.AllowRead, false),
				AllowWrite:           valueOr(spec.AllowWrite, false),
				AllowUnauthenticated: allowUnauthenticated,
				IsPublic:             isPublic,
				CronSchedule:         spec.CronSchedule,
				CreatedBy:            createdBy,
			}

			if err := h.storage.CreateFunction(ctx, fn); err != nil {
				errorList = append(errorList, fiber.Map{
					"function": spec.Name,
					"error":    err.Error(),
					"action":   "create",
				})
				continue
			}

			created = append(created, spec.Name)
		}
	}

	// Delete removed functions (after successful creates/updates for safety)
	if req.Options.DeleteMissing {
		for _, name := range toDelete {
			if err := h.storage.DeleteFunctionForSync(syncCtx, name, namespace, currentTenantID); err != nil {
				errorList = append(errorList, fiber.Map{
					"function": name,
					"error":    err.Error(),
					"action":   "delete",
				})
				continue
			}
			deleted = append(deleted, name)
		}
	}

	return c.JSON(fiber.Map{
		"message":   "Functions synced successfully",
		"namespace": namespace,
		"summary": fiber.Map{
			"created":   len(created),
			"updated":   len(updated),
			"deleted":   len(deleted),
			"unchanged": len(unchanged),
			"errors":    len(errorList),
		},
		"details": fiber.Map{
			"created":   created,
			"updated":   updated,
			"deleted":   deleted,
			"unchanged": unchanged,
		},
		"errors":  errorList,
		"dry_run": false,
	})
}

// LoadFromFilesystem loads functions from filesystem at boot time
// This is called from main.go if auto_load_on_boot is enabled
func (h *Handler) LoadFromFilesystem(ctx context.Context) error {
	// Scan functions directory for all .ts files
	functionFiles, err := ListFunctionFiles(h.functionsDir)
	if err != nil {
		return fmt.Errorf("failed to scan functions directory: %w", err)
	}

	// Track results
	created := []string{}
	updated := []string{}
	errors := []string{}

	// Process each function file
	for _, fileInfo := range functionFiles {
		// Check if function exists in database
		existingFn, err := h.storage.GetFunction(ctx, fileInfo.Name)

		if err != nil {
			// Function doesn't exist in database - create it
			bundledCode, originalCode, isBundled, bundleError, err := h.bundleFunctionFromFilesystem(ctx, fileInfo.Name)
			if err != nil {
				errors = append(errors, fmt.Sprintf("%s: %v", fileInfo.Name, err))
				continue
			}

			// Parse configuration from code comments
			config := ParseFunctionConfig(originalCode)

			// Create new function with default settings
			fn := &EdgeFunction{
				Name:                 fileInfo.Name,
				Code:                 bundledCode,
				OriginalCode:         &originalCode,
				IsBundled:            isBundled,
				BundleError:          bundleError,
				Enabled:              true,
				TimeoutSeconds:       30,
				MemoryLimitMB:        128,
				AllowNet:             true,
				AllowEnv:             true,
				AllowRead:            false,
				AllowWrite:           false,
				AllowUnauthenticated: config.AllowUnauthenticated,
				IsPublic:             config.IsPublic,
				DisableExecutionLogs: config.DisableExecutionLogs,
				Source:               "filesystem",
			}

			if err := h.storage.CreateFunction(ctx, fn); err != nil {
				errors = append(errors, fmt.Sprintf("%s: failed to create: %v", fileInfo.Name, err))
				continue
			}

			created = append(created, fileInfo.Name)
		} else {
			// Function exists - update code from filesystem
			bundledCode, originalCode, isBundled, bundleError, err := h.bundleFunctionFromFilesystem(ctx, fileInfo.Name)
			if err != nil {
				errors = append(errors, fmt.Sprintf("%s: %v", fileInfo.Name, err))
				continue
			}

			// Parse configuration from code comments
			config := ParseFunctionConfig(originalCode)

			// Update if code or config has changed
			// Compare with original_code if available, otherwise with code
			compareCode := originalCode
			if existingFn.OriginalCode != nil {
				compareCode = *existingFn.OriginalCode
			}

			if existingFn.Code != bundledCode || compareCode != originalCode || existingFn.AllowUnauthenticated != config.AllowUnauthenticated || existingFn.IsPublic != config.IsPublic || existingFn.DisableExecutionLogs != config.DisableExecutionLogs {
				updates := map[string]interface{}{
					"code":                   bundledCode,
					"original_code":          originalCode,
					"is_bundled":             isBundled,
					"bundle_error":           bundleError,
					"allow_unauthenticated":  config.AllowUnauthenticated,
					"is_public":              config.IsPublic,
					"disable_execution_logs": config.DisableExecutionLogs,
				}

				if err := h.storage.UpdateFunction(ctx, fileInfo.Name, updates); err != nil {
					errors = append(errors, fmt.Sprintf("%s: failed to update: %v", fileInfo.Name, err))
					continue
				}

				updated = append(updated, fileInfo.Name)
			}
		}
	}

	// Note: Auto-load does NOT delete functions missing from filesystem
	// This prevents data loss when UI-created functions exist alongside file-based functions
	// Use the manual reload endpoint to perform full sync including deletions

	// Log results
	if len(created) > 0 || len(updated) > 0 {
		fmt.Printf("Functions loaded from filesystem: %d created, %d updated\n", len(created), len(updated))
	}
	if len(errors) > 0 {
		fmt.Printf("Errors loading functions: %v\n", errors)
	}

	return nil
}

// CreateSharedModule creates a new shared module
func (h *Handler) CreateSharedModule(c fiber.Ctx) error {
	var req struct {
		ModulePath  string  `json:"module_path"`
		Content     string  `json:"content"`
		Description *string `json:"description"`
	}

	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Validate module_path starts with _shared/
	if !strings.HasPrefix(req.ModulePath, "_shared/") {
		return c.Status(400).JSON(fiber.Map{"error": "Module path must start with '_shared/'"})
	}

	// Get user ID from context (if authenticated)
	var userID *uuid.UUID
	if uid := c.Locals("user_id"); uid != nil {
		if parsedUID, ok := uid.(uuid.UUID); ok {
			userID = &parsedUID
		}
	}

	module := &SharedModule{
		ModulePath:  req.ModulePath,
		Content:     req.Content,
		Description: req.Description,
		CreatedBy:   userID,
	}

	if err := h.storage.CreateSharedModule(middleware.CtxWithTenant(c), module); err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "already exists") {
			return c.Status(409).JSON(fiber.Map{"error": "Shared module already exists"})
		}
		log.Error().Err(err).Str("module_path", req.ModulePath).Msg("Failed to create shared module")
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create shared module"})
	}

	log.Info().
		Str("module_path", module.ModulePath).
		Str("user_id", toString(userID)).
		Msg("Shared module created")

	return c.Status(201).JSON(module)
}

// ListSharedModules returns all shared modules
func (h *Handler) ListSharedModules(c fiber.Ctx) error {
	modules, err := h.storage.ListSharedModules(middleware.CtxWithTenant(c))
	if err != nil {
		log.Error().Err(err).Msg("Failed to list shared modules")
		return c.Status(500).JSON(fiber.Map{"error": "Failed to list shared modules"})
	}

	return c.JSON(modules)
}

// GetSharedModule retrieves a shared module by path
func (h *Handler) GetSharedModule(c fiber.Ctx) error {
	// Get full path from wildcard (e.g., "cors.ts" from "/shared/cors.ts")
	modulePath := strings.TrimPrefix(c.Params("*"), "/")

	// Ensure it starts with _shared/
	if !strings.HasPrefix(modulePath, "_shared/") {
		modulePath = "_shared/" + modulePath
	}

	module, err := h.storage.GetSharedModule(middleware.CtxWithTenant(c), modulePath)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return c.Status(404).JSON(fiber.Map{"error": "Shared module not found"})
		}
		log.Error().Err(err).Str("module_path", modulePath).Msg("Failed to get shared module")
		return c.Status(500).JSON(fiber.Map{"error": "Failed to get shared module"})
	}

	return c.JSON(module)
}

// UpdateSharedModule updates an existing shared module
func (h *Handler) UpdateSharedModule(c fiber.Ctx) error {
	// Get full path from wildcard
	modulePath := strings.TrimPrefix(c.Params("*"), "/")

	// Ensure it starts with _shared/
	if !strings.HasPrefix(modulePath, "_shared/") {
		modulePath = "_shared/" + modulePath
	}

	var req struct {
		Content     string  `json:"content"`
		Description *string `json:"description"`
	}

	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if err := h.storage.UpdateSharedModule(middleware.CtxWithTenant(c), modulePath, req.Content, req.Description); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return c.Status(404).JSON(fiber.Map{"error": "Shared module not found"})
		}
		log.Error().Err(err).Str("module_path", modulePath).Msg("Failed to update shared module")
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update shared module"})
	}

	// Get updated module
	module, err := h.storage.GetSharedModule(middleware.CtxWithTenant(c), modulePath)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Module updated but failed to retrieve"})
	}

	log.Info().
		Str("module_path", modulePath).
		Int("version", module.Version).
		Msg("Shared module updated")

	return c.JSON(module)
}

// DeleteSharedModule deletes a shared module
func (h *Handler) DeleteSharedModule(c fiber.Ctx) error {
	// Get full path from wildcard
	modulePath := strings.TrimPrefix(c.Params("*"), "/")

	// Ensure it starts with _shared/
	if !strings.HasPrefix(modulePath, "_shared/") {
		modulePath = "_shared/" + modulePath
	}

	if err := h.storage.DeleteSharedModule(middleware.CtxWithTenant(c), modulePath); err != nil {
		log.Error().Err(err).Str("module_path", modulePath).Msg("Failed to delete shared module")
		return c.Status(500).JSON(fiber.Map{"error": "Failed to delete shared module"})
	}

	log.Info().Str("module_path", modulePath).Msg("Shared module deleted")

	return c.JSON(fiber.Map{"message": "Shared module deleted successfully"})
}
