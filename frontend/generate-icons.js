// Simple script to generate PWA icons using macOS sips
// Run with: node generate-icons.js

import fs from 'fs';
import path from 'path';
import { execSync } from 'child_process';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

// Create a simple SVG icon with the app's gradient theme
const createIconSVG = (size) => {
  return `<svg width="${size}" height="${size}" viewBox="0 0 ${size} ${size}" xmlns="http://www.w3.org/2000/svg">
  <defs>
    <linearGradient id="grad" x1="0%" y1="0%" x2="100%" y2="100%">
      <stop offset="0%" style="stop-color:#60a5fa;stop-opacity:1" />
      <stop offset="50%" style="stop-color:#22d3ee;stop-opacity:1" />
      <stop offset="100%" style="stop-color:#a78bfa;stop-opacity:1" />
    </linearGradient>
  </defs>
  <rect width="${size}" height="${size}" rx="${size * 0.2}" fill="url(#grad)"/>
  <text x="50%" y="50%" font-family="Arial, sans-serif" font-size="${size * 0.3}" font-weight="bold" fill="white" text-anchor="middle" dominant-baseline="central">TT</text>
</svg>`;
};

const publicDir = path.join(__dirname, 'public');

if (!fs.existsSync(publicDir)) {
  fs.mkdirSync(publicDir, { recursive: true });
}

// Create temporary SVG files
const svg192 = path.join(publicDir, 'icon-192-temp.svg');
const svg512 = path.join(publicDir, 'icon-512-temp.svg');
const png192 = path.join(publicDir, 'icon-192.png');
const png512 = path.join(publicDir, 'icon-512.png');

fs.writeFileSync(svg192, createIconSVG(192));
fs.writeFileSync(svg512, createIconSVG(512));

// Try to convert using qlmanage (macOS built-in) or provide instructions
try {
  // Use qlmanage to convert SVG to PNG (macOS)
  console.log('Converting SVG to PNG...');
  execSync(`qlmanage -t -s 192 -o ${publicDir} ${svg192}`, { stdio: 'ignore' });
  execSync(`qlmanage -t -s 512 -o ${publicDir} ${svg512}`, { stdio: 'ignore' });
  
  // Rename files (qlmanage outputs with .png extension)
  if (fs.existsSync(path.join(publicDir, 'icon-192-temp.svg.png'))) {
    fs.renameSync(path.join(publicDir, 'icon-192-temp.svg.png'), png192);
  }
  if (fs.existsSync(path.join(publicDir, 'icon-512-temp.svg.png'))) {
    fs.renameSync(path.join(publicDir, 'icon-512-temp.svg.png'), png512);
  }
  
  // Clean up temp files
  fs.unlinkSync(svg192);
  fs.unlinkSync(svg512);
  
  console.log('✅ Icons created successfully!');
} catch (error) {
  console.log('⚠️  Could not convert SVG to PNG automatically.');
  console.log('SVG files created in public/ directory.');
  console.log('Please convert them to PNG manually or use an online tool.');
  console.log('You can use: https://realfavicongenerator.net/');
}

