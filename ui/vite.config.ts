import { defineConfig } from 'vite'
import react, { reactCompilerPreset } from '@vitejs/plugin-react'
import babel from '@rolldown/plugin-babel'
import { resolve } from 'node:path'

// https://vite.dev/config/
export default defineConfig({
  base: './',
  build: {
    outDir: 'dist/renderer',
    rolldownOptions: {
      input: {
        // 主窗口入口
        main: resolve(import.meta.dirname, './main/index.html'),
        // 托盘弹出窗口入口
        popup: resolve(import.meta.dirname, './popup/index.html'), 
      },
    },
  },
  server: {
    middlewareMode: false,
    port: 5173,
  },
  plugins: [
    react(),
    babel({ presets: [reactCompilerPreset()] })
  ],
})
