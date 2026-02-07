import { defineConfig } from 'astro/config'

export default defineConfig({
  output: 'static',
  site: 'https://kpm.fyi',
  server: {
    host: '0.0.0.0',
    port: 4321,
  },
})
