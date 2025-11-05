// Simple script to generate PWA icons using macOS sips
// Run with: node generate-icons.js

import fs from 'fs';
import path from 'path';
import { execSync } from 'child_process';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

// Create a turn tracker icon with circular arrow representing turns
const createIconSVG = (size) => {
  const scale = size / 512;
  const center = size / 2;
  const radius = size * 0.43;
  const strokeWidth = size * 0.055;
  const arrowSize = size * 0.04;
  
  return `<svg width="${size}" height="${size}" viewBox="0 0 ${size} ${size}" xmlns="http://www.w3.org/2000/svg">
  <defs>
    <linearGradient id="bgGrad${size}" x1="0%" y1="0%" x2="100%" y2="100%">
      <stop offset="0%" style="stop-color:#3b82f6;stop-opacity:1" />
      <stop offset="50%" style="stop-color:#06b6d4;stop-opacity:1" />
      <stop offset="100%" style="stop-color:#8b5cf6;stop-opacity:1" />
    </linearGradient>
    <linearGradient id="arrowGrad${size}" x1="0%" y1="0%" x2="100%" y2="100%">
      <stop offset="0%" style="stop-color:#ffffff;stop-opacity:1" />
      <stop offset="100%" style="stop-color:#e0e7ff;stop-opacity:1" />
    </linearGradient>
  </defs>
  
  <!-- Background with rounded corners -->
  <rect width="${size}" height="${size}" rx="${size * 0.2}" fill="url(#bgGrad${size})"/>
  
  <!-- Outer ring -->
  <circle cx="${center}" cy="${center}" r="${radius}" fill="none" stroke="rgba(255,255,255,0.2)" stroke-width="${strokeWidth * 0.3}"/>
  
  <!-- Circular arrow representing turn rotation -->
  <g transform="translate(${center}, ${center})">
    <!-- Circular arrow arc -->
    <path d="M 0,-${radius} A ${radius},${radius} 0 1,1 0,${radius}" 
          fill="none" 
          stroke="url(#arrowGrad${size})" 
          stroke-width="${strokeWidth}" 
          stroke-linecap="round"
          stroke-dasharray="${Math.PI * radius * 2} ${Math.PI * radius * 2}"
          stroke-dashoffset="${Math.PI * radius * 0.5}"/>
    
    <!-- Arrow head -->
    <polygon points="-${arrowSize},-${radius} 0,-${radius - arrowSize} ${arrowSize},-${radius} 0,-${radius + arrowSize}" 
             fill="url(#arrowGrad${size})"/>
    
    <!-- Center circle -->
    <circle cx="0" cy="0" r="${size * 0.12}" fill="rgba(255,255,255,0.15)"/>
    <circle cx="0" cy="0" r="${size * 0.1}" fill="rgba(255,255,255,0.25)"/>
  </g>
</svg>`;
};

const publicDir = path.join(__dirname, 'public');

if (!fs.existsSync(publicDir)) {
  fs.mkdirSync(publicDir, { recursive: true });
}

// Create SVG files
const svgFile = path.join(publicDir, 'icon.svg');
const svg192 = path.join(publicDir, 'icon-192-temp.svg');
const svg512 = path.join(publicDir, 'icon-512-temp.svg');
const png192 = path.join(publicDir, 'icon-192.png');
const png512 = path.join(publicDir, 'icon-512.png');

// Create main SVG (512x512)
const mainSVG = createIconSVG(512);
fs.writeFileSync(svgFile, mainSVG);

// Create size-specific SVGs for conversion
fs.writeFileSync(svg192, createIconSVG(192));
fs.writeFileSync(svg512, createIconSVG(512));

console.log('📝 SVG files created');

// Try multiple conversion methods
let converted = false;

// Method 1: Try using sips (macOS built-in, more reliable)
try {
  console.log('Trying sips conversion...');
  execSync(`sips -s format png -z 192 192 ${svg192} --out ${png192}`, { stdio: 'ignore' });
  execSync(`sips -s format png -z 512 512 ${svg512} --out ${png512}`, { stdio: 'ignore' });
  
  if (fs.existsSync(png192) && fs.existsSync(png512)) {
    converted = true;
    console.log('✅ Icons converted using sips');
  }
} catch (error) {
  // Continue to next method
}

// Method 2: Try using qlmanage (macOS built-in)
if (!converted) {
  try {
    console.log('Trying qlmanage conversion...');
    execSync(`qlmanage -t -s 192 -o ${publicDir} ${svg192}`, { stdio: 'ignore' });
    execSync(`qlmanage -t -s 512 -o ${publicDir} ${svg512}`, { stdio: 'ignore' });
    
    // Rename files (qlmanage outputs with .png extension)
    const ql192 = path.join(publicDir, 'icon-192-temp.svg.png');
    const ql512 = path.join(publicDir, 'icon-512-temp.svg.png');
    
    if (fs.existsSync(ql192)) {
      fs.renameSync(ql192, png192);
    }
    if (fs.existsSync(ql512)) {
      fs.renameSync(ql512, png512);
    }
    
    if (fs.existsSync(png192) && fs.existsSync(png512)) {
      converted = true;
      console.log('✅ Icons converted using qlmanage');
    }
  } catch (error) {
    // Continue
  }
}

// Clean up temp files
if (fs.existsSync(svg192)) {
  fs.unlinkSync(svg192);
}
if (fs.existsSync(svg512)) {
  fs.unlinkSync(svg512);
}

if (converted) {
  console.log('✅ All icons created successfully!');
  console.log(`   - ${svgFile}`);
  console.log(`   - ${png192}`);
  console.log(`   - ${png512}`);
} else {
  console.log('⚠️  Could not convert SVG to PNG automatically.');
  console.log('✅ SVG file created:', svgFile);
  console.log('📝 Please convert SVG to PNG manually:');
  console.log('   - Open icon.svg in a browser or image editor');
  console.log('   - Export/resize to 192x192 and 512x512 PNG');
  console.log('   - Or use an online tool like https://realfavicongenerator.net/');
}

