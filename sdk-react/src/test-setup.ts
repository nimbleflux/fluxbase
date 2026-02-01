/**
 * Test setup file for Fluxbase React SDK
 */

import { vi } from 'vitest';
import '@testing-library/react';
import '@testing-library/jest-dom/vitest';

// Mock window.location for tests that need it
Object.defineProperty(window, 'location', {
  value: {
    href: 'http://localhost:3000',
    origin: 'http://localhost:3000',
    pathname: '/',
    search: '',
    hash: '',
    assign: vi.fn(),
    replace: vi.fn(),
    reload: vi.fn(),
  },
  writable: true,
});
