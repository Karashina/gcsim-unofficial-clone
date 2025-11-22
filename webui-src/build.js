const esbuild = require('esbuild');
const path = require('path');
const fs = require('fs');

// Build JavaScript
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
  console.log('✓ JavaScript built:', path.resolve(__dirname, '..', 'webui', 'app.js'));
}).catch((e) => {
  console.error('JavaScript build failed:', e);
  process.exit(1);
});

// Copy CSS (simple concatenation for now)
const cssFiles = [
  'src/styles/results.css'
];

const cssOutput = path.resolve(__dirname, '..', 'webui', 'results.css');
let cssContent = '/* Generated CSS Bundle - DO NOT EDIT DIRECTLY */\n\n';

cssFiles.forEach(file => {
  const filePath = path.resolve(__dirname, file);
  if (fs.existsSync(filePath)) {
    cssContent += `/* From: ${file} */\n`;
    cssContent += fs.readFileSync(filePath, 'utf8');
    cssContent += '\n\n';
  }
});

fs.writeFileSync(cssOutput, cssContent);
console.log('✓ CSS built:', cssOutput);
