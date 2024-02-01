import { nodeResolve } from "@rollup/plugin-node-resolve"
import terser from '@rollup/plugin-terser'

export default {
  input: "./lua-editor.mjs",
  output: {
    file: 'lua-bundle.min.js',
    format: 'iife',
    name: 'version',
    // plugins: [terser()]
  },
  plugins: [nodeResolve()]
}
