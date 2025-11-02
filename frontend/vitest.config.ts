import { defineConfig } from 'vitest/config';
import react from '@vitejs/plugin-react';
import path from 'path';

export default defineConfig({
  plugins: [react()],
  test: {
    environment: 'jsdom',
    globals: true,
    setupFiles: ['./vitest.setup.ts'],
    exclude: [
      '**/node_modules/**', 
      '**/e2e/**', 
      '**/.{idea,git,cache,output,temp}/**', 
      '**/build/**',
      '**/.react-router/**', // Add this to exclude React Router build files
    ],
    // Add resolve conditions to prevent accessing build artifacts
    resolve: {
      conditions: ['import', 'module', 'browser', 'default'],
    },
  },
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
    // Prevent React Router from trying to access build files
    conditions: ['import', 'module', 'browser', 'default'],
  },
});

