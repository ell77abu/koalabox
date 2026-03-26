#!/bin/bash
# 快速部署脚本
# 使用方法: bash deploy.sh

echo "========================================="
echo "    开始构建和部署前端应用"
echo "========================================="

# 检查是否安装了 Node.js 和 npm
if ! command -v npm &> /dev/null; then
    echo "❌ 错误: 未找到 npm，请先安装 Node.js"
    exit 1
fi

echo ""
echo "📦 Step 1: 安装依赖..."
npm install

echo ""
echo "🔨 Step 2: 生产构建..."
npm run build

if [ $? -ne 0 ]; then
    echo "❌ 构建失败"
    exit 1
fi

echo ""
echo "✅ 构建成功！"
echo ""
echo "📁 产物位置: ./dist"
echo ""
echo "🚀 部署步骤:"
echo "   1. 修改 .env.production 中的服务器地址"
echo "   2. 将 dist 目录上传到服务器"
echo "   3. 配置 Web 服务器（Nginx 或 Apache）"
echo "   4. 配置 API 代理指向后端服务"
echo ""
echo "📖 详细说明请查看 DEPLOYMENT.md"
echo ""
echo "========================================="
