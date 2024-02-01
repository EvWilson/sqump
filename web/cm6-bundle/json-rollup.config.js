import { nodeResolve } from "@rollup/plugin-node-resolve"
import terser from '@rollup/plugin-terser'

export default {
  input: "./json-editor.mjs",
  output: {
    file: 'json-bundle.min.js',
    format: 'iife',
    name: 'version',
    // plugins: [terser()]
  },
  plugins: [nodeResolve()]
}
