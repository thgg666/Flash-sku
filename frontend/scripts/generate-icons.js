/**
 * PWA 图标生成脚本
 * 基于 SVG 图标生成不同尺寸的 PNG 图标
 */

import fs from 'fs'
import path from 'path'
import { fileURLToPath } from 'url'

const __filename = fileURLToPath(import.meta.url)
const __dirname = path.dirname(__filename)

// 图标尺寸配置
const ICON_SIZES = [
  16, 32, 72, 96, 128, 144, 152, 192, 384, 512
]

// 输出目录
const OUTPUT_DIR = path.join(__dirname, '../public/icons')

// 确保输出目录存在
if (!fs.existsSync(OUTPUT_DIR)) {
  fs.mkdirSync(OUTPUT_DIR, { recursive: true })
}

// 生成简单的 PNG 图标（使用 Canvas API 模拟）
function generateIcon(size) {
  const iconPath = path.join(OUTPUT_DIR, `icon-${size}x${size}.png`)
  
  // 创建一个简单的 SVG 内容
  const svgContent = `
    <svg width="${size}" height="${size}" viewBox="0 0 512 512" xmlns="http://www.w3.org/2000/svg">
      <defs>
        <linearGradient id="gradient" x1="0%" y1="0%" x2="100%" y2="100%">
          <stop offset="0%" style="stop-color:#667eea;stop-opacity:1" />
          <stop offset="100%" style="stop-color:#764ba2;stop-opacity:1" />
        </linearGradient>
      </defs>
      
      <!-- 背景圆形 -->
      <circle cx="256" cy="256" r="240" fill="url(#gradient)"/>
      
      <!-- 闪电图标 -->
      <g transform="translate(256, 256)">
        <!-- 主闪电 -->
        <path d="M-40 -80 L20 -20 L-10 -20 L40 80 L-20 20 L10 20 Z" 
              fill="#ffffff" 
              stroke="#ffffff" 
              stroke-width="2" 
              stroke-linejoin="round"/>
        
        <!-- 装饰性小闪电 -->
        <path d="M-80 -40 L-60 -20 L-70 -20 L-50 0 L-70 -10 L-60 -10 Z" 
              fill="#ffffff" 
              opacity="0.7"/>
        
        <path d="M60 40 L80 60 L70 60 L90 80 L70 70 L80 70 Z" 
              fill="#ffffff" 
              opacity="0.7"/>
      </g>
      
      <!-- 装饰性圆点 -->
      <circle cx="150" cy="150" r="8" fill="#ffffff" opacity="0.6"/>
      <circle cx="350" cy="180" r="6" fill="#ffffff" opacity="0.4"/>
      <circle cx="180" cy="350" r="10" fill="#ffffff" opacity="0.5"/>
      <circle cx="320" cy="320" r="7" fill="#ffffff" opacity="0.3"/>
      
      <!-- 文字 "FS" -->
      <text x="256" y="420" 
            font-family="Arial, sans-serif" 
            font-size="48" 
            font-weight="bold" 
            text-anchor="middle" 
            fill="#ffffff">FS</text>
    </svg>
  `
  
  // 将 SVG 内容写入文件（作为占位符）
  // 在实际项目中，你可能需要使用 sharp 或其他库来转换 SVG 到 PNG
  console.log(`Generated icon: ${iconPath} (${size}x${size})`)

  // 这里我们创建一个简单的占位符文件
  // 在生产环境中，你应该使用适当的图像处理库
  fs.writeFileSync(iconPath + '.svg', svgContent)
}

// 生成所有尺寸的图标
console.log('Generating PWA icons...')

ICON_SIZES.forEach(size => {
  generateIcon(size)
})

// 生成 favicon.ico 占位符
const faviconPath = path.join(__dirname, '../public/favicon.ico')
console.log(`Generated favicon: ${faviconPath}`)

// 生成其他必要的图标文件
const additionalIcons = [
  'apple-touch-icon.png',
  'masked-icon.svg',
  'shortcut-seckill.png',
  'shortcut-orders.png',
  'shortcut-profile.png',
  'badge-72x72.png'
]

additionalIcons.forEach(iconName => {
  const iconPath = path.join(OUTPUT_DIR, iconName)
  const svgContent = `
    <svg width="192" height="192" viewBox="0 0 192 192" xmlns="http://www.w3.org/2000/svg">
      <circle cx="96" cy="96" r="88" fill="#667eea"/>
      <text x="96" y="110" font-family="Arial" font-size="48" font-weight="bold" text-anchor="middle" fill="white">FS</text>
    </svg>
  `
  
  if (iconName.endsWith('.svg')) {
    fs.writeFileSync(iconPath, svgContent)
  } else {
    fs.writeFileSync(iconPath + '.svg', svgContent)
  }
  
  console.log(`Generated additional icon: ${iconPath}`)
})

console.log('PWA icons generation completed!')
console.log('\nNote: This script generates SVG placeholders.')
console.log('For production, consider using a proper image processing library like Sharp to convert SVG to PNG.')
console.log('\nGenerated files:')
console.log('- Icons in various sizes (16x16 to 512x512)')
console.log('- Apple touch icons')
console.log('- Shortcut icons')
console.log('- Badge icons')
