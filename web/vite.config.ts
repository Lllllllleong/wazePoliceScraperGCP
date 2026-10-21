/// <reference types="vitest/config" />
import { defineConfig } from 'vite';
import { svelte } from '@sveltejs/vite-plugin-svelte';

export default defineConfig({
  plugins: [svelte()],
  build: {
    target: 'es2022',
    // MapLibre alone is ~800 kB minified.
    chunkSizeWarningLimit: 1200,
  },
  test: {
    include: ['src/**/*.test.ts'],
  },
});
