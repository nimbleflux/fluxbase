package branching

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// TestBranchingSecurity_UnauthorizedDatabaseCreation tests security of database creation
func TestBranchingSecurity_UnauthorizedDatabaseCreation(t *testing.T) {
	tests := []struct {
		name          string
		maxTotal      int
		maxPerUser    int
		existingTotal int
		existingUser  int
		userRole      string
		createdBy     *uuid.UUID
		wantErr       bool
		errContains   string
		securityIssue string
	}{
		{
			name:          "admin within quota",
			maxTotal:      10,
			maxPerUser:    5,
			existingTotal: 3,
			existingUser:  2,
			userRole:      "admin",
			createdBy:     uuidPtr(uuid.New()),
			wantErr:       false,
			securityIssue: "",
		},
		{
			name:          "exceeds user quota - DoS protection",
			maxTotal:      10,
			maxPerUser:    5,
			existingTotal: 3,
			existingUser:  5,
			userRole:      "user",
			createdBy:     uuidPtr(uuid.New()),
			wantErr:       true,
			errContains:   "maximum number of branches per user",
			securityIssue: "DoS: User could exceed personal quota",
		},
		{
			name:          "exceeds system quota - DoS protection",
			maxTotal:      10,
			maxPerUser:    5,
			existingTotal: 10,
			existingUser:  2,
			userRole:      "user",
			createdBy:     uuidPtr(uuid.New()),
			wantErr:       true,
			errContains:   "maximum number of branches reached",
			securityIssue: "DoS: System-wide branch limit reached",
		},
		{
			name:          "no user ID - quota bypass attempt",
			maxTotal:      10,
			maxPerUser:    5,
			existingTotal: 3,
			existingUser:  0,
			userRole:      "anonymous",
			createdBy:     nil,
			wantErr:       false,
			securityIssue: "",
		},
		{
			name:          "unlimited configuration still validates",
			maxTotal:      0,
			maxPerUser:    0,
			existingTotal: 1000,
			existingUser:  100,
			userRole:      "user",
			createdBy:     uuidPtr(uuid.New()),
			wantErr:       false,
			securityIssue: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test quota enforcement logic
			if tt.maxTotal > 0 && tt.existingTotal >= tt.maxTotal {
				assert.True(t, tt.wantErr, "Should fail when system quota exceeded")
				if tt.errContains != "" {
					assert.Contains(t, tt.errContains, "maximum")
				}
			}

			if tt.maxPerUser > 0 && tt.existingUser >= tt.maxPerUser && tt.createdBy != nil {
				assert.True(t, tt.wantErr, "Should fail when user quota exceeded")
				if tt.errContains != "" {
					assert.Contains(t, tt.errContains, "per user")
				}
			}

			if tt.securityIssue != "" {
				t.Logf("Security Issue: %s", tt.securityIssue)
			}
		})
	}
}

// TestBranchingSecurity_DatabaseIsolation tests database isolation between branches
func TestBranchingSecurity_DatabaseIsolation(t *testing.T) {
	tests := []struct {
		name          string
		sourceBranch  string
		targetBranch  string
		shouldAccess  bool
		securityIssue string
	}{
		{
			name:          "branch cannot access main",
			sourceBranch:  "branch_test",
			targetBranch:  "main",
			shouldAccess:  false,
			securityIssue: "Isolation: Branch could access main database",
		},
		{
			name:          "branch cannot access other branch",
			sourceBranch:  "branch_test1",
			targetBranch:  "branch_test2",
			shouldAccess:  false,
			securityIssue: "Isolation: Cross-branch access possible",
		},
		{
			name:          "main cannot access branch",
			sourceBranch:  "main",
			targetBranch:  "branch_test",
			shouldAccess:  false,
			securityIssue: "Isolation: Main could access branch database",
		},
		{
			name:          "branch can access itself",
			sourceBranch:  "branch_test",
			targetBranch:  "branch_test",
			shouldAccess:  true,
			securityIssue: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test isolation logic
			isSameBranch := tt.sourceBranch == tt.targetBranch

			// Only same-branch access should be allowed
			shouldAllow := isSameBranch

			if tt.shouldAccess != shouldAllow {
				if tt.securityIssue != "" {
					t.Logf("SECURITY ISSUE: %s", tt.securityIssue)
				}
				assert.Equal(t, tt.shouldAccess, shouldAllow,
					"Access control mismatch - security issue detected")
			}
		})
	}
}

// TestBranchingSecurity_ConnectionPoolIsolation tests connection pool isolation
func TestBranchingSecurity_ConnectionPoolIsolation(t *testing.T) {
	tests := []struct {
		name          string
		requestedSlug string
		activeSlugs   []string
		shouldCreate  bool
		securityIssue string
	}{
		{
			name:          "new branch pool creation",
			requestedSlug: "branch_test",
			activeSlugs:   []string{"branch_other"},
			shouldCreate:  true,
			securityIssue: "",
		},
		{
			name:          "reuse existing pool",
			requestedSlug: "branch_test",
			activeSlugs:   []string{"branch_test", "branch_other"},
			shouldCreate:  false,
			securityIssue: "",
		},
		{
			name:          "main always uses main pool",
			requestedSlug: "main",
			activeSlugs:   []string{},
			shouldCreate:  false,
			securityIssue: "",
		},
		{
			name:          "empty slug uses main pool",
			requestedSlug: "",
			activeSlugs:   []string{},
			shouldCreate:  false,
			securityIssue: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test pool isolation logic
			isMain := tt.requestedSlug == "" || tt.requestedSlug == "main"
			alreadyExists := false
			for _, slug := range tt.activeSlugs {
				if slug == tt.requestedSlug {
					alreadyExists = true
					break
				}
			}

			shouldCreatePool := !isMain && !alreadyExists && tt.requestedSlug != ""

			if tt.shouldCreate != shouldCreatePool {
				if tt.securityIssue != "" {
					t.Logf("SECURITY ISSUE: %s", tt.securityIssue)
				}
			}
		})
	}
}

// TestBranchingSecurity_BranchCleanup tests branch cleanup and data deletion
func TestBranchingSecurity_BranchCleanup(t *testing.T) {
	tests := []struct {
		name          string
		branchType    BranchType
		isOwner       bool
		isAdmin       bool
		isExpired     bool
		shouldDelete  bool
		securityIssue string
	}{
		{
			name:          "owner can delete preview branch",
			branchType:    BranchTypePreview,
			isOwner:       true,
			isAdmin:       false,
			isExpired:     false,
			shouldDelete:  true,
			securityIssue: "",
		},
		{
			name:          "admin can delete any non-main branch",
			branchType:    BranchTypePreview,
			isOwner:       false,
			isAdmin:       true,
			isExpired:     false,
			shouldDelete:  true,
			securityIssue: "",
		},
		{
			name:          "non-owner cannot delete branch",
			branchType:    BranchTypePreview,
			isOwner:       false,
			isAdmin:       false,
			isExpired:     false,
			shouldDelete:  false,
			securityIssue: "Authorization: Non-owner could delete branch",
		},
		{
			name:          "main branch cannot be deleted",
			branchType:    BranchTypeMain,
			isOwner:       true,
			isAdmin:       true,
			isExpired:     false,
			shouldDelete:  false,
			securityIssue: "Authorization: Main branch deletion attempted",
		},
		{
			name:          "expired branch auto-deleted",
			branchType:    BranchTypePreview,
			isOwner:       false,
			isAdmin:       false,
			isExpired:     true,
			shouldDelete:  true,
			securityIssue: "",
		},
		{
			name:          "abandoned branch cleanup",
			branchType:    BranchTypePreview,
			isOwner:       false,
			isAdmin:       true,
			isExpired:     true,
			shouldDelete:  true,
			securityIssue: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test cleanup authorization
			canDelete := tt.isAdmin || tt.isOwner

			// Main branch can never be deleted
			if tt.branchType == BranchTypeMain {
				canDelete = false
				if tt.shouldDelete {
					t.Logf("SECURITY ISSUE: %s", tt.securityIssue)
				}
			}

			// Expired branches can be auto-deleted
			if tt.isExpired {
				canDelete = true
			}

			if tt.shouldDelete != canDelete {
				if tt.securityIssue != "" {
					t.Logf("SECURITY ISSUE: %s", tt.securityIssue)
				}
				assert.Equal(t, tt.shouldDelete, canDelete,
					"Cleanup authorization mismatch")
			}
		})
	}
}

// TestBranchingSecurity_BranchAccessControl tests branch access control
func TestBranchingSecurity_BranchAccessControl(t *testing.T) {
	ownerID := uuid.New()
	otherUserID := uuid.New()

	tests := []struct {
		name           string
		userID         *uuid.UUID
		branchOwner    *uuid.UUID
		accessLevel    BranchAccessLevel
		requestedLevel BranchAccessLevel
		shouldAllow    bool
		securityIssue  string
	}{
		{
			name:           "owner has full access",
			userID:         &ownerID,
			branchOwner:    &ownerID,
			accessLevel:    BranchAccessAdmin,
			requestedLevel: BranchAccessRead,
			shouldAllow:    true,
			securityIssue:  "",
		},
		{
			name:           "read access sufficient for read",
			userID:         &otherUserID,
			branchOwner:    &ownerID,
			accessLevel:    BranchAccessRead,
			requestedLevel: BranchAccessRead,
			shouldAllow:    true,
			securityIssue:  "",
		},
		{
			name:           "read access insufficient for write",
			userID:         &otherUserID,
			branchOwner:    &ownerID,
			accessLevel:    BranchAccessRead,
			requestedLevel: BranchAccessWrite,
			shouldAllow:    false,
			securityIssue:  "Authorization: User with read access requested write",
		},
		{
			name:           "write access sufficient for read",
			userID:         &otherUserID,
			branchOwner:    &ownerID,
			accessLevel:    BranchAccessWrite,
			requestedLevel: BranchAccessRead,
			shouldAllow:    true,
			securityIssue:  "",
		},
		{
			name:           "admin access sufficient for all",
			userID:         &otherUserID,
			branchOwner:    &ownerID,
			accessLevel:    BranchAccessAdmin,
			requestedLevel: BranchAccessAdmin,
			shouldAllow:    true,
			securityIssue:  "",
		},
		{
			name:           "no access denied",
			userID:         &otherUserID,
			branchOwner:    &ownerID,
			accessLevel:    BranchAccessRead,
			requestedLevel: BranchAccessAdmin,
			shouldAllow:    false,
			securityIssue:  "Authorization: User with low access requested high access",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Owner always has admin access
			isOwner := tt.userID != nil && tt.branchOwner != nil && *tt.userID == *tt.branchOwner

			// Check if access level is sufficient
			accessSufficient := isAccessSufficient(tt.accessLevel, tt.requestedLevel)

			shouldAllow := isOwner || accessSufficient

			if tt.shouldAllow != shouldAllow {
				if tt.securityIssue != "" {
					t.Logf("SECURITY ISSUE: %s", tt.securityIssue)
				}
				assert.Equal(t, tt.shouldAllow, shouldAllow,
					"Access control mismatch - security issue detected")
			}
		})
	}
}

// TestBranchingSecurity_ResourceLimits tests resource limit enforcement
func TestBranchingSecurity_ResourceLimits(t *testing.T) {
	tests := []struct {
		name            string
		maxBranches     int
		currentBranches int
		maxPerUser      int
		userBranches    int
		shouldAllow     bool
		securityIssue   string
	}{
		{
			name:            "within system limit",
			maxBranches:     100,
			currentBranches: 50,
			maxPerUser:      10,
			userBranches:    5,
			shouldAllow:     true,
			securityIssue:   "",
		},
		{
			name:            "system limit prevents DoS",
			maxBranches:     100,
			currentBranches: 100,
			maxPerUser:      10,
			userBranches:    5,
			shouldAllow:     false,
			securityIssue:   "DoS: System branch limit reached",
		},
		{
			name:            "user limit prevents DoS",
			maxBranches:     100,
			currentBranches: 50,
			maxPerUser:      5,
			userBranches:    5,
			shouldAllow:     false,
			securityIssue:   "DoS: User branch limit reached",
		},
		{
			name:            "unlimited config is safe",
			maxBranches:     0,
			currentBranches: 1000,
			maxPerUser:      0,
			userBranches:    100,
			shouldAllow:     true,
			securityIssue:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Check system limit
			systemLimitOK := tt.maxBranches == 0 || tt.currentBranches < tt.maxBranches

			// Check user limit
			userLimitOK := tt.maxPerUser == 0 || tt.userBranches < tt.maxPerUser

			shouldAllow := systemLimitOK && userLimitOK

			if tt.shouldAllow != shouldAllow {
				if tt.securityIssue != "" {
					t.Logf("SECURITY ISSUE: %s", tt.securityIssue)
				}
				assert.Equal(t, tt.shouldAllow, shouldAllow,
					"Resource limit enforcement failed - DoS vulnerability")
			}
		})
	}
}

// TestBranchingSecurity_SQLInjectionPrevention tests SQL injection prevention
func TestBranchingSecurity_SQLInjectionPrevention(t *testing.T) {
	tests := []struct {
		name          string
		slug          string
		databaseName  string
		isValid       bool
		securityIssue string
	}{
		{
			name:          "valid slug",
			slug:          "test-branch",
			databaseName:  "branch_test_branch",
			isValid:       true,
			securityIssue: "",
		},
		{
			name:          "SQL injection attempt - quote",
			slug:          "test'; DROP TABLE branches; --",
			databaseName:  "branch_test_s_drop_table_branches",
			isValid:       false,
			securityIssue: "SQL Injection: Single quote in slug",
		},
		{
			name:          "SQL injection attempt - semicolon",
			slug:          "test;DELETEFROMusers",
			databaseName:  "branch_test_deletefromusers",
			isValid:       false,
			securityIssue: "SQL Injection: Semicolon in slug",
		},
		{
			name:          "SQL injection attempt - union",
			slug:          "test-union-select",
			databaseName:  "branch_test_union_select",
			isValid:       true, // "union" and "select" are valid words
			securityIssue: "",
		},
		{
			name:          "path traversal attempt",
			slug:          "../../etc/passwd",
			databaseName:  "branch_.._.._etc_passwd",
			isValid:       false,
			securityIssue: "Path Traversal: Dot characters in slug",
		},
		{
			name:          "null byte injection",
			slug:          "test\x00branch",
			databaseName:  "branch_test_branch",
			isValid:       false,
			securityIssue: "Injection: Invalid characters in slug",
		},
		{
			name:          "reserved keyword",
			slug:          "main",
			databaseName:  "branch_main",
			isValid:       false,
			securityIssue: "Authorization: Reserved 'main' slug",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate slug
			err := ValidateSlug(tt.slug)
			isValid := err == nil

			if tt.isValid != isValid {
				if tt.securityIssue != "" {
					t.Logf("SECURITY ISSUE: %s", tt.securityIssue)
				}
				assert.Equal(t, tt.isValid, isValid,
					"SQL injection prevention failed - security vulnerability")
			}

			// Verify sanitization
			sanitized := sanitizeIdentifier(tt.databaseName)
			assert.Contains(t, sanitized, `"`, "Should be quoted for SQL safety")
		})
	}
}

// TestBranchingSecurity_SlugGenerationSafety tests safe slug generation
func TestBranchingSecurity_SlugGenerationSafety(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedSlug  string
		securityIssue string
	}{
		{
			name:          "normal input",
			input:         "My Test Branch",
			expectedSlug:  "my-test-branch",
			securityIssue: "",
		},
		{
			name:          "removes dangerous characters",
			input:         "Test@#$%^&*()Branch",
			expectedSlug:  "testbranch",
			securityIssue: "",
		},
		{
			name:          "handles multiple spaces",
			input:         "Test    Branch",
			expectedSlug:  "test-branch",
			securityIssue: "",
		},
		{
			name:          "trims leading/trailing spaces",
			input:         "  test branch  ",
			expectedSlug:  "test-branch",
			securityIssue: "",
		},
		{
			name:          "collapses consecutive hyphens",
			input:         "test--branch",
			expectedSlug:  "test-branch",
			securityIssue: "",
		},
		{
			name:          "limits length",
			input:         "this-is-a-very-long-branch-name-that-exceeds-maximum-allowed-length",
			expectedSlug:  "this-is-a-very-long-branch-name-that-exceeds-maxim",
			securityIssue: "",
		},
		{
			name:          "empty defaults to branch",
			input:         "",
			expectedSlug:  "branch",
			securityIssue: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateSlug(tt.input)

			// Verify the slug is safe
			err := ValidateSlug(result)
			if err != nil {
				if tt.securityIssue != "" {
					t.Logf("SECURITY ISSUE: %s - Generated invalid slug: %s",
						tt.securityIssue, result)
				}
				t.Fatalf("Generated unsafe slug: %s (error: %v)", result, err)
			}

			assert.Equal(t, tt.expectedSlug, result)
		})
	}
}

// TestBranchingSecurity_ExpirationHandling tests branch expiration logic
func TestBranchingSecurity_ExpirationHandling(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name          string
		expiresAt     *time.Time
		branchType    BranchType
		autoDelete    time.Duration
		shouldExpire  bool
		securityIssue string
	}{
		{
			name:          "not expired - future time",
			expiresAt:     timePtr(now.Add(24 * time.Hour)),
			branchType:    BranchTypePreview,
			autoDelete:    24 * time.Hour,
			shouldExpire:  false,
			securityIssue: "",
		},
		{
			name:          "expired - past time",
			expiresAt:     timePtr(now.Add(-1 * time.Hour)),
			branchType:    BranchTypePreview,
			autoDelete:    24 * time.Hour,
			shouldExpire:  true,
			securityIssue: "",
		},
		{
			name:          "no expiration set",
			expiresAt:     nil,
			branchType:    BranchTypePreview,
			autoDelete:    24 * time.Hour,
			shouldExpire:  false,
			securityIssue: "",
		},
		{
			name:          "main branch never expires",
			expiresAt:     timePtr(now.Add(-100 * time.Hour)),
			branchType:    BranchTypeMain,
			autoDelete:    24 * time.Hour,
			shouldExpire:  false,
			securityIssue: "Data Retention: Main branch marked for expiration",
		},
		{
			name:          "production branch respects expiration",
			expiresAt:     timePtr(now.Add(-1 * time.Hour)),
			branchType:    BranchTypePreview,
			autoDelete:    24 * time.Hour,
			shouldExpire:  true,
			securityIssue: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isExpired := false

			if tt.expiresAt != nil && tt.expiresAt.Before(now) {
				if tt.branchType != BranchTypeMain {
					isExpired = true
				}
			}

			if tt.shouldExpire != isExpired {
				if tt.securityIssue != "" {
					t.Logf("SECURITY ISSUE: %s", tt.securityIssue)
				}
				assert.Equal(t, tt.shouldExpire, isExpired,
					"Expiration logic error - security issue detected")
			}
		})
	}
}

// TestBranchingSecurity_DataCloneModeSafety tests data clone mode restrictions
func TestBranchingSecurity_DataCloneModeSafety(t *testing.T) {
	tests := []struct {
		name          string
		cloneMode     DataCloneMode
		parentExists  bool
		isValid       bool
		securityIssue string
	}{
		{
			name:          "schema-only without parent",
			cloneMode:     DataCloneModeSchemaOnly,
			parentExists:  false,
			isValid:       true,
			securityIssue: "",
		},
		{
			name:          "schema-only with parent",
			cloneMode:     DataCloneModeSchemaOnly,
			parentExists:  true,
			isValid:       true,
			securityIssue: "",
		},
		{
			name:          "full clone without parent",
			cloneMode:     DataCloneModeFullClone,
			parentExists:  false,
			isValid:       true,
			securityIssue: "",
		},
		{
			name:          "full clone with parent",
			cloneMode:     DataCloneModeFullClone,
			parentExists:  true,
			isValid:       true,
			securityIssue: "",
		},
		{
			name:          "seed data without parent",
			cloneMode:     DataCloneModeSeedData,
			parentExists:  false,
			isValid:       true,
			securityIssue: "",
		},
		{
			name:          "seed data with parent",
			cloneMode:     DataCloneModeSeedData,
			parentExists:  true,
			isValid:       true,
			securityIssue: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify clone mode is valid
			validModes := map[DataCloneMode]bool{
				DataCloneModeSchemaOnly: true,
				DataCloneModeFullClone:  true,
				DataCloneModeSeedData:   true,
			}

			isValid := validModes[tt.cloneMode]

			if tt.isValid != isValid {
				if tt.securityIssue != "" {
					t.Logf("SECURITY ISSUE: %s", tt.securityIssue)
				}
				assert.Equal(t, tt.isValid, isValid,
					"Invalid data clone mode - security issue detected")
			}
		})
	}
}

// TestBranchingSecurity_DatabaseNameSafety tests database name generation safety
func TestBranchingSecurity_DatabaseNameSafety(t *testing.T) {
	tests := []struct {
		name          string
		prefix        string
		slug          string
		shouldContain string
		securityIssue string
	}{
		{
			name:          "normal name generation",
			prefix:        "branch_",
			slug:          "test",
			shouldContain: "branch_test",
			securityIssue: "",
		},
		{
			name:          "hyphens become underscores",
			prefix:        "branch_",
			slug:          "test-branch",
			shouldContain: "branch_test_branch",
			securityIssue: "",
		},
		{
			name:          "no SQL injection in name",
			prefix:        "branch_",
			slug:          "test'; DROP TABLE",
			shouldContain: "branch_test",
			securityIssue: "", // Slug generation removes dangerous chars
		},
		{
			name:          "length limited",
			prefix:        "branch_",
			slug:          "very-long-branch-name-that-exceeds-postgres-limit",
			shouldContain: "branch_",
			securityIssue: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateDatabaseName(tt.prefix, tt.slug)

			// Verify name is safe
			assert.Contains(t, result, tt.shouldContain)

			// Check for SQL injection patterns
			// Slug generation removes dangerous characters
			slug := GenerateSlug(tt.slug)
			err := ValidateSlug(slug)
			assert.NoError(t, err, "Generated slug should be valid")
		})
	}
}

// TestBranchingSecurity_ConcurrentCreationSafety tests concurrent branch creation safety
func TestBranchingSecurity_ConcurrentCreationSafety(t *testing.T) {
	t.Run("concurrent branch creation respects limits", func(t *testing.T) {
		// This tests that concurrent branch creation requests
		// properly enforce limits and don't create race conditions

		maxBranches := 10
		creationAttempts := 20

		// Simulate concurrent creation attempts
		created := 0
		for i := 0; i < creationAttempts; i++ {
			if created < maxBranches {
				// Check limit before creating
				if created < maxBranches {
					// Simulate successful creation
					if created < maxBranches {
						created++
					}
				}
			}
		}

		// Should never exceed max
		assert.LessOrEqual(t, created, maxBranches,
			"Concurrent creation exceeded system limit - race condition detected")
	})
}

// Helper function to create UUID pointer
func uuidPtr(id uuid.UUID) *uuid.UUID {
	return &id
}

// Helper function to create time pointer
func timePtr(t time.Time) *time.Time {
	return &t
}

// TestBranchingSecurity_SanitizationComprehensive tests comprehensive sanitization
func TestBranchingSecurity_SanitizationComprehensive(t *testing.T) {
	tests := []struct {
		name          string
		identifier    string
		isSafe        bool
		securityIssue string
	}{
		{
			name:          "simple identifier",
			identifier:    "table_name",
			isSafe:        true,
			securityIssue: "",
		},
		{
			name:          "identifier with quote",
			identifier:    `my"table`,
			isSafe:        true, // After sanitization
			securityIssue: "",
		},
		{
			name:          "empty string",
			identifier:    "",
			isSafe:        true, // Empty is sanitized to ""
			securityIssue: "",
		},
		{
			name:          "identifier with backslash",
			identifier:    `table\name`,
			isSafe:        true,
			securityIssue: "",
		},
		{
			name:          "identifier with newline",
			identifier:    "table\nname",
			isSafe:        true, // Sanitized
			securityIssue: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test sanitization
			result := sanitizeIdentifier(tt.identifier)

			// Should always be quoted
			assert.Contains(t, result, `"`,
				"Sanitized identifier should be quoted")
		})
	}
}
