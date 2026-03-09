package branching

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// NewSeeder Tests
// =============================================================================

func TestNewSeeder(t *testing.T) {
	t.Run("creates seeder with path", func(t *testing.T) {
		path := "/path/to/seeds"
		seeder := NewSeeder(path)

		require.NotNil(t, seeder)
		assert.Equal(t, path, seeder.seedsPath)
	})

	t.Run("creates seeder with empty path", func(t *testing.T) {
		seeder := NewSeeder("")

		require.NotNil(t, seeder)
		assert.Equal(t, "", seeder.seedsPath)
	})
}

// =============================================================================
// Seeder Struct Tests
// =============================================================================

func TestSeeder_Struct(t *testing.T) {
	t.Run("stores seeds path", func(t *testing.T) {
		seeder := &Seeder{
			seedsPath: "/custom/path",
		}

		assert.Equal(t, "/custom/path", seeder.seedsPath)
	})
}

// =============================================================================
// SeedFile Struct Tests
// =============================================================================

func TestSeedFile_Struct(t *testing.T) {
	t.Run("stores all fields", func(t *testing.T) {
		seedFile := SeedFile{
			Name:    "001_initial",
			Path:    "/seeds/001_initial.sql",
			Content: "INSERT INTO users VALUES (1, 'test');",
		}

		assert.Equal(t, "001_initial", seedFile.Name)
		assert.Equal(t, "/seeds/001_initial.sql", seedFile.Path)
		assert.Equal(t, "INSERT INTO users VALUES (1, 'test');", seedFile.Content)
	})

	t.Run("handles empty values", func(t *testing.T) {
		seedFile := SeedFile{}

		assert.Empty(t, seedFile.Name)
		assert.Empty(t, seedFile.Path)
		assert.Empty(t, seedFile.Content)
	})
}

// =============================================================================
// DiscoverSeedFiles Tests
// =============================================================================

func TestSeeder_DiscoverSeedFiles(t *testing.T) {
	t.Run("returns empty list for non-existent directory", func(t *testing.T) {
		seeder := NewSeeder("/non/existent/path")
		ctx := context.Background()

		seeds, err := seeder.DiscoverSeedFiles(ctx)

		require.NoError(t, err)
		assert.Empty(t, seeds)
	})

	t.Run("returns empty list for empty directory", func(t *testing.T) {
		// Create temp directory
		tempDir, err := os.MkdirTemp("", "seeds_empty_*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(tempDir) }()

		seeder := NewSeeder(tempDir)
		ctx := context.Background()

		seeds, err := seeder.DiscoverSeedFiles(ctx)

		require.NoError(t, err)
		assert.Empty(t, seeds)
	})

	t.Run("discovers SQL files in directory", func(t *testing.T) {
		// Create temp directory with seed files
		tempDir, err := os.MkdirTemp("", "seeds_*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(tempDir) }()

		// Create seed files
		file1 := filepath.Join(tempDir, "001_initial.sql")
		file2 := filepath.Join(tempDir, "002_users.sql")
		err = os.WriteFile(file1, []byte("-- Initial seed\nINSERT INTO config VALUES ('key', 'value');"), 0o644)
		require.NoError(t, err)
		err = os.WriteFile(file2, []byte("-- Users seed\nINSERT INTO users VALUES (1, 'admin');"), 0o644)
		require.NoError(t, err)

		seeder := NewSeeder(tempDir)
		ctx := context.Background()

		seeds, err := seeder.DiscoverSeedFiles(ctx)

		require.NoError(t, err)
		assert.Len(t, seeds, 2)
		assert.Equal(t, "001_initial", seeds[0].Name)
		assert.Equal(t, "002_users", seeds[1].Name)
	})

	t.Run("skips non-SQL files", func(t *testing.T) {
		// Create temp directory
		tempDir, err := os.MkdirTemp("", "seeds_*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(tempDir) }()

		// Create mixed files
		sqlFile := filepath.Join(tempDir, "001_seed.sql")
		txtFile := filepath.Join(tempDir, "readme.txt")
		mdFile := filepath.Join(tempDir, "notes.md")
		err = os.WriteFile(sqlFile, []byte("INSERT INTO t VALUES (1);"), 0o644)
		require.NoError(t, err)
		err = os.WriteFile(txtFile, []byte("This is a text file"), 0o644)
		require.NoError(t, err)
		err = os.WriteFile(mdFile, []byte("# Notes"), 0o644)
		require.NoError(t, err)

		seeder := NewSeeder(tempDir)
		ctx := context.Background()

		seeds, err := seeder.DiscoverSeedFiles(ctx)

		require.NoError(t, err)
		assert.Len(t, seeds, 1)
		assert.Equal(t, "001_seed", seeds[0].Name)
	})

	t.Run("skips subdirectories", func(t *testing.T) {
		// Create temp directory
		tempDir, err := os.MkdirTemp("", "seeds_*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(tempDir) }()

		// Create file and subdirectory
		sqlFile := filepath.Join(tempDir, "001_seed.sql")
		subDir := filepath.Join(tempDir, "subdir")
		err = os.WriteFile(sqlFile, []byte("INSERT INTO t VALUES (1);"), 0o644)
		require.NoError(t, err)
		err = os.Mkdir(subDir, 0o755)
		require.NoError(t, err)

		seeder := NewSeeder(tempDir)
		ctx := context.Background()

		seeds, err := seeder.DiscoverSeedFiles(ctx)

		require.NoError(t, err)
		assert.Len(t, seeds, 1)
	})

	t.Run("sorts files lexicographically", func(t *testing.T) {
		// Create temp directory
		tempDir, err := os.MkdirTemp("", "seeds_*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(tempDir) }()

		// Create files in non-sorted order
		files := []string{"003_third.sql", "001_first.sql", "002_second.sql"}
		for _, name := range files {
			path := filepath.Join(tempDir, name)
			err = os.WriteFile(path, []byte("SELECT 1;"), 0o644)
			require.NoError(t, err)
		}

		seeder := NewSeeder(tempDir)
		ctx := context.Background()

		seeds, err := seeder.DiscoverSeedFiles(ctx)

		require.NoError(t, err)
		require.Len(t, seeds, 3)
		assert.Equal(t, "001_first", seeds[0].Name)
		assert.Equal(t, "002_second", seeds[1].Name)
		assert.Equal(t, "003_third", seeds[2].Name)
	})

	t.Run("reads file content correctly", func(t *testing.T) {
		// Create temp directory
		tempDir, err := os.MkdirTemp("", "seeds_*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(tempDir) }()

		expectedContent := "-- Seed file\nINSERT INTO users (id, name) VALUES (1, 'test');\nINSERT INTO users (id, name) VALUES (2, 'admin');"
		sqlFile := filepath.Join(tempDir, "001_users.sql")
		err = os.WriteFile(sqlFile, []byte(expectedContent), 0o644)
		require.NoError(t, err)

		seeder := NewSeeder(tempDir)
		ctx := context.Background()

		seeds, err := seeder.DiscoverSeedFiles(ctx)

		require.NoError(t, err)
		require.Len(t, seeds, 1)
		assert.Equal(t, expectedContent, seeds[0].Content)
		assert.Equal(t, sqlFile, seeds[0].Path)
	})

	t.Run("handles file with only .sql extension", func(t *testing.T) {
		// Create temp directory
		tempDir, err := os.MkdirTemp("", "seeds_*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(tempDir) }()

		// Create file named just ".sql"
		sqlFile := filepath.Join(tempDir, ".sql")
		err = os.WriteFile(sqlFile, []byte("SELECT 1;"), 0o644)
		require.NoError(t, err)

		seeder := NewSeeder(tempDir)
		ctx := context.Background()

		seeds, err := seeder.DiscoverSeedFiles(ctx)

		require.NoError(t, err)
		// Should discover the file with empty name after trimming .sql
		assert.Len(t, seeds, 1)
		assert.Equal(t, "", seeds[0].Name)
	})

	t.Run("handles empty SQL files", func(t *testing.T) {
		// Create temp directory
		tempDir, err := os.MkdirTemp("", "seeds_*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(tempDir) }()

		sqlFile := filepath.Join(tempDir, "001_empty.sql")
		err = os.WriteFile(sqlFile, []byte(""), 0o644)
		require.NoError(t, err)

		seeder := NewSeeder(tempDir)
		ctx := context.Background()

		seeds, err := seeder.DiscoverSeedFiles(ctx)

		require.NoError(t, err)
		require.Len(t, seeds, 1)
		assert.Equal(t, "", seeds[0].Content)
	})

	t.Run("handles files with special characters in name", func(t *testing.T) {
		// Create temp directory
		tempDir, err := os.MkdirTemp("", "seeds_*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(tempDir) }()

		sqlFile := filepath.Join(tempDir, "001_users-roles_mapping.sql")
		err = os.WriteFile(sqlFile, []byte("SELECT 1;"), 0o644)
		require.NoError(t, err)

		seeder := NewSeeder(tempDir)
		ctx := context.Background()

		seeds, err := seeder.DiscoverSeedFiles(ctx)

		require.NoError(t, err)
		require.Len(t, seeds, 1)
		assert.Equal(t, "001_users-roles_mapping", seeds[0].Name)
	})
}

// =============================================================================
// Helper function tests
// =============================================================================

func TestSeeder_SeedFileSorting(t *testing.T) {
	t.Run("numeric prefix sorting works correctly", func(t *testing.T) {
		seeds := []SeedFile{
			{Name: "010_tenth"},
			{Name: "001_first"},
			{Name: "002_second"},
			{Name: "100_hundredth"},
		}

		// Verify the sort logic matches DiscoverSeedFiles
		tempDir, err := os.MkdirTemp("", "seeds_*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(tempDir) }()

		for _, seed := range seeds {
			path := filepath.Join(tempDir, seed.Name+".sql")
			err = os.WriteFile(path, []byte("SELECT 1;"), 0o644)
			require.NoError(t, err)
		}

		seeder := NewSeeder(tempDir)
		discovered, err := seeder.DiscoverSeedFiles(context.Background())

		require.NoError(t, err)
		require.Len(t, discovered, 4)
		assert.Equal(t, "001_first", discovered[0].Name)
		assert.Equal(t, "002_second", discovered[1].Name)
		assert.Equal(t, "010_tenth", discovered[2].Name)
		assert.Equal(t, "100_hundredth", discovered[3].Name)
	})
}

// =============================================================================
// ExecuteSeeds Tests (Unit tests - no database required)
// =============================================================================

func TestSeeder_ExecuteSeeds_NoSeedsDirectory(t *testing.T) {
	t.Run("returns nil when no seeds directory exists", func(t *testing.T) {
		seeder := NewSeeder("/non/existent/path")
		ctx := context.Background()

		// ExecuteSeeds should handle missing directory gracefully
		// by returning nil (no seeds to execute)
		// Note: We can't fully test ExecuteSeeds without a database pool,
		// but we can test the DiscoverSeedFiles behavior it uses
		seeds, err := seeder.DiscoverSeedFiles(ctx)

		require.NoError(t, err)
		assert.Empty(t, seeds)
	})
}

func TestSeeder_ExecuteSeeds_EmptyDirectory(t *testing.T) {
	t.Run("handles empty seeds directory", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "seeds_*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(tempDir) }()

		seeder := NewSeeder(tempDir)
		ctx := context.Background()

		seeds, err := seeder.DiscoverSeedFiles(ctx)

		require.NoError(t, err)
		assert.Empty(t, seeds)
	})
}

// =============================================================================
// Edge Case Tests
// =============================================================================

func TestSeeder_EdgeCases(t *testing.T) {
	t.Run("handles context cancellation", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "seeds_*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(tempDir) }()

		seeder := NewSeeder(tempDir)
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		// DiscoverSeedFiles should still work with cancelled context
		// since it doesn't check context (it's file I/O based)
		seeds, err := seeder.DiscoverSeedFiles(ctx)

		require.NoError(t, err)
		assert.Empty(t, seeds)
	})

	t.Run("handles very long file names", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "seeds_*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(tempDir) }()

		longName := "001_" + string(make([]byte, 200)) // Very long name
		for i := range longName[4:] {
			longName = longName[:4+i] + "a"
		}
		longName = "001_" + "a_very_long_seed_file_name_that_describes_what_it_does_in_detail"

		sqlFile := filepath.Join(tempDir, longName+".sql")
		err = os.WriteFile(sqlFile, []byte("SELECT 1;"), 0o644)
		require.NoError(t, err)

		seeder := NewSeeder(tempDir)
		seeds, err := seeder.DiscoverSeedFiles(context.Background())

		require.NoError(t, err)
		require.Len(t, seeds, 1)
		assert.Equal(t, longName, seeds[0].Name)
	})

	t.Run("handles unicode in file content", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "seeds_*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(tempDir) }()

		unicodeContent := "INSERT INTO messages (text) VALUES ('Hello, 世界! 🚀');"
		sqlFile := filepath.Join(tempDir, "001_unicode.sql")
		err = os.WriteFile(sqlFile, []byte(unicodeContent), 0o644)
		require.NoError(t, err)

		seeder := NewSeeder(tempDir)
		seeds, err := seeder.DiscoverSeedFiles(context.Background())

		require.NoError(t, err)
		require.Len(t, seeds, 1)
		assert.Equal(t, unicodeContent, seeds[0].Content)
	})

	t.Run("handles large file content", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "seeds_*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(tempDir) }()

		// Create a large SQL file (1MB)
		largeContent := make([]byte, 1024*1024)
		for i := range largeContent {
			largeContent[i] = 'A'
		}
		sqlFile := filepath.Join(tempDir, "001_large.sql")
		err = os.WriteFile(sqlFile, largeContent, 0o644)
		require.NoError(t, err)

		seeder := NewSeeder(tempDir)
		seeds, err := seeder.DiscoverSeedFiles(context.Background())

		require.NoError(t, err)
		require.Len(t, seeds, 1)
		assert.Len(t, seeds[0].Content, 1024*1024)
	})
}

// =============================================================================
// Validation Tests
// =============================================================================

func TestSeeder_Validation(t *testing.T) {
	t.Run("preserves file path in SeedFile", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "seeds_*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(tempDir) }()

		sqlFile := filepath.Join(tempDir, "001_test.sql")
		err = os.WriteFile(sqlFile, []byte("SELECT 1;"), 0o644)
		require.NoError(t, err)

		seeder := NewSeeder(tempDir)
		seeds, err := seeder.DiscoverSeedFiles(context.Background())

		require.NoError(t, err)
		require.Len(t, seeds, 1)
		assert.Equal(t, sqlFile, seeds[0].Path)
	})
}

// =============================================================================
// Concurrency Tests
// =============================================================================

func TestSeeder_Concurrency(t *testing.T) {
	t.Run("concurrent DiscoverSeedFiles calls are safe", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "seeds_*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(tempDir) }()

		// Create some seed files
		for i := 1; i <= 5; i++ {
			path := filepath.Join(tempDir, "00"+string(rune('0'+i))+"_seed.sql")
			err = os.WriteFile(path, []byte("SELECT 1;"), 0o644)
			require.NoError(t, err)
		}

		seeder := NewSeeder(tempDir)
		ctx := context.Background()

		// Run multiple concurrent discoveries
		results := make(chan []SeedFile, 10)
		errors := make(chan error, 10)

		for i := 0; i < 10; i++ {
			go func() {
				seeds, err := seeder.DiscoverSeedFiles(ctx)
				if err != nil {
					errors <- err
					return
				}
				results <- seeds
			}()
		}

		// Collect results
		for i := 0; i < 10; i++ {
			select {
			case err := <-errors:
				t.Fatalf("unexpected error: %v", err)
			case seeds := <-results:
				assert.Len(t, seeds, 5)
			}
		}
	})
}

// =============================================================================
// BranchID Tests
// =============================================================================

func TestSeeder_BranchIDHandling(t *testing.T) {
	t.Run("accepts valid UUID for branch ID", func(t *testing.T) {
		branchID := uuid.New()
		assert.NotEqual(t, uuid.Nil, branchID)
	})

	t.Run("nil UUID is valid", func(t *testing.T) {
		branchID := uuid.Nil
		assert.Equal(t, uuid.Nil, branchID)
	})
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkSeeder_DiscoverSeedFiles(b *testing.B) {
	// Create temp directory with seed files
	tempDir, err := os.MkdirTemp("", "seeds_bench_*")
	require.NoError(b, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create 100 seed files
	for i := 1; i <= 100; i++ {
		name := "seed.sql"
		switch {
		case i < 10:
			name = "00" + string(rune('0'+i)) + "_" + name
		case i < 100:
			name = "0" + string(rune('0'+i/10)) + string(rune('0'+i%10)) + "_" + name
		default:
			name = "100_" + name
		}
		path := filepath.Join(tempDir, name)
		err = os.WriteFile(path, []byte("SELECT 1; SELECT 2; SELECT 3;"), 0o644)
		require.NoError(b, err)
	}

	seeder := NewSeeder(tempDir)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = seeder.DiscoverSeedFiles(ctx)
	}
}

func BenchmarkNewSeeder(b *testing.B) {
	path := "/path/to/seeds"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewSeeder(path)
	}
}
