import { defineConfig } from 'astro/config'

export default defineConfig({
  site: 'https://kpm.fyi',
  output: 'static',
  server: {
    host: '0.0.0.0',
    port: 4321,
  },
})
