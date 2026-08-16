import { build as viteBuild } from 'vite'
import { build as esbuild } from 'esbuild'

console.log('Building Renderer...')
await viteBuild({ configFile: 'vite.config.ts' })

console.log('Building Main...')
await esbuild({
    entryPoints: ['electron/main.ts'],
    outfile: 'dist/electron/main.js',
    bundle: true,
    platform: 'node',
    format: 'esm',
    external: ['electron'],
    minify: true, // 生产环境开启压缩
  })

console.log('Building Preload...')
await esbuild({
    entryPoints: ['electron/preload.ts'],
    outfile: 'dist/electron/preload.js',
    bundle: true,
    platform: 'node',
    format: 'cjs',
    external: ['electron'],
    minify: true,
  })

console.log('Build finished successfully!')