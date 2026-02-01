import { defineConfig } from 'vitest/config';
import { rmSync } from 'fs';
import { join } from 'path';

// Clear Vitest cache when config loads
try {
  const cacheDir = join(process.cwd(), 'node_modules', '.vitest');
  rmSync(cacheDir, { recursive: true, force: true });
  console.log('Cleared Vitest cache directory');
} catch (err) {
  // Ignore if cache doesn't exist
}

export default defineConfig({
  test: {
    cache: false,
    coverage: {
      provider: 'v8',
      reporter: ['text', 'json', 'html', 'lcov'],
      reportsDirectory: './coverage',
      include: ['src/**/*.ts'],
      exclude: ['src/**/*.test.ts', 'src/**/*.d.ts', 'src/examples/**'],
      thresholds: {
        statements: 50,
        branches: 50,
        functions: 50,
        lines: 50
      }
    }
  }
});
