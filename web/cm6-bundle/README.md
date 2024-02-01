# Code Mirror 6 Bundler

This is largely just taken from the [official Code Mirror bundling example](https://codemirror.net/examples/bundle/) with some light additions from the [Rollup tutorial](https://rollupjs.org/tutorial/#using-output-plugins) to add minification.

All that should ever need to be run is `npm install && npm run build` to create the minified bundle and move it into the `assets` folder to be included in the final binary.
