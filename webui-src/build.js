const esbuild = require('esbuild');
const path = require('path');

esbuild.build({
  entryPoints: ['src/app.js'],
  bundle: true,
  outfile: path.resolve(__dirname, '..', 'webui', 'app.js'),
  platform: 'browser',
  format: 'iife',
  target: ['es2017'],
  minify: false,
  sourcemap: false,
}).then(() => {
  console.log('Build completed:', path.resolve(__dirname, '..', 'webui', 'app.js'));
}).catch((e) => {
  console.error(e);
  process.exit(1);
});
