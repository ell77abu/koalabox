# 前端部署指南

本文档说明如何将 Vue 前端应用部署到服务器。

## 📌 目录结构

```
lab1-frontend/
├── src/                     # 源代码
├── dist/                    # 生产构建输出（部署时使用）
├── .env.development         # 开发环境配置（本地 Mock）
├── .env.production          # 生产环境配置（真实后端）
├── vite.config.mjs          # Vite 构建配置
├── package.json             # 项目配置和依赖
└── README.md                # 本文件
```

---

## 🛠️ 环境配置说明

### **开发环境** (.env.development)
```
VITE_API_BASE_URL=http://localhost:3001  # Mock 后端地址
VITE_API_PREFIX=/api/v1
```

### **生产环境** (.env.production)
```
VITE_API_BASE_URL=http://your-server-ip:8080  # 改为你的服务器地址
VITE_API_PREFIX=/api/v1
```

#### 常见服务器地址配置:
- **内网服务器**: `http://192.168.1.100:8080`
- **外网服务器**: `http://api.yourdomain.com` 或 `https://api.yourdomain.com`
- **云服务器**: `http://your-cloud-ip:8080`

**⚠️ 重要**: 修改 `.env.production` 中的 `VITE_API_BASE_URL` 为你的真实后端服务器地址。

---

## 📦 构建流程

### **1. 本地开发（使用 Mock）**
```bash
npm run dev:all
# 或
npm run dev:mock &  # 启动 Mock 后端
npm run dev         # 启动前端开发服务器
```

### **2. 生产构建**
```bash
npm run build
```

执行此命令会：
- 编译 Vue 组件到 JavaScript
- 压缩和优化代码
- 输出到 `dist/` 目录
- **系统自动使用 `.env.production` 配置**

构建完成后，`dist/` 目录结构：
```
dist/
├── index.html          # 入口 HTML 文件
├── assets/
│   ├── index-xxx.js    # 打包后的 JavaScript（哈希文件名）
│   ├── index-xxx.css   # 打包后的 CSS（哈希文件名）
│   └── ...
```

### **3. 本地测试生产构建**
```bash
npm run preview
# 访问 http://localhost:4173 查看生产包效果
```

**⚠️ 注意**: `preview` 命令在本地测试，仍然需要 Mock 服务器运行或改为指向真实后端。

---

## 🚀 部署到服务器

### **方案 A：部署到 Nginx（推荐）**

#### **Step 1: 本地构建**
```bash
npm run build
```

#### **Step 2: 上传到服务器**
```bash
# 将 dist 目录上传到服务器（例如 /var/www/html/lab1-frontend）
scp -r dist/ user@your-server:/var/www/html/lab1-frontend/
```

#### **Step 3: 配置 Nginx**
编辑 `/etc/nginx/sites-available/default`（或你的 Nginx 配置文件）：

```nginx
server {
    listen 80;
    server_name your-domain.com;  # 或服务器的 IP 地址

    root /var/www/html/lab1-frontend;  # dist 目录路径
    index index.html;

    location / {
        # 支持 SPA 路由：所有请求都指向 index.html
        try_files $uri $uri/ /index.html;
    }

    # API 代理（可选，如果前端和后端不在同一域名）
    location /api/v1 {
        proxy_pass http://127.0.0.1:8080;  # 后端服务器地址
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

#### **Step 4: 重启 Nginx**
```bash
sudo nginx -t              # 检查配置是否正确
sudo systemctl restart nginx  # 重启 Nginx
```

---

### **方案 B：部署到 Apache**

编辑 `.htaccess`（在 `dist/` 目录中）：

```apache
<IfModule mod_rewrite.c>
  RewriteEngine On
  RewriteBase /
  RewriteRule ^index\.html$ - [L]
  RewriteCond %{REQUEST_FILENAME} !-f
  RewriteCond %{REQUEST_FILENAME} !-d
  RewriteRule . /index.html [L]
</IfModule>
```

上传 `dist/` 目录及其 `.htaccess` 文件到服务器的 Web 根目录。

---

### **方案 C：部署到容器（Docker）**

如果使用 Docker（可选），创建 `Dockerfile`：

```dockerfile
# Stage 1: Build
FROM node:18-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm install
COPY . .
RUN npm run build

# Stage 2: Serve with Nginx
FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

创建 `nginx.conf`：
```nginx
server {
    listen 80;
    location / {
        root /usr/share/nginx/html;
        try_files $uri $uri/ /index.html;
    }
    location /api/v1 {
        proxy_pass http://your-backend-service:8080;
    }
}
```

构建和运行：
```bash
docker build -t lab1-frontend .
docker run -p 80:80 lab1-frontend
```

---

## ✅ 检查清单

部署前，确保：

- [ ] **修改 `.env.production`**: 将 `VITE_API_BASE_URL` 改为你的真实后端地址
- [ ] **运行 `npm run build`**: 生成 `dist/` 目录
- [ ] **测试构建**: 运行 `npm run preview` 验证生产包
- [ ] **上传到服务器**: 将 `dist/` 目录上传
- [ ] **配置 Web 服务器**: Nginx / Apache / 其他
- [ ] **配置 API 代理**: 如果前后端不同域
- [ ] **测试路由**: 刷新页面确保 SPA 路由正常工作
- [ ] **检查浏览器控制台**: 确认没有 CORS 错误
- [ ] **验证 API 请求**: 在浏览器 DevTools 的 Network 查看请求是否到达真实后端

---

## 🔍 常见问题

### **Q1: 如何判断前端使用的是哪个环境配置？**

检查浏览器开发者工具的 Network 标签，查看 API 请求的完整 URL：
- **开发**: `http://localhost:5173/api/v1/...` 
- **生产**: `http://your-server:8080/api/v1/...`

### **Q2: 部署后出现 404 错误**

**原因**: 没有配置 SPA 路由回退。  
**解决**: 配置 Web 服务器的 `try_files` / `.htaccess`，使所有路由都指向 `index.html`。

### **Q3: 出现 CORS 错误**

**原因**: 前端和后端的域名/端口不同。  
**解决**:
- 在后端 API 配置 CORS 头
- 或在 Web 服务器（Nginx）配置 API 代理

### **Q4: 如何在生产环境测试 Mock 数据？**

修改 `.env.production`，临时改回 Mock 地址：
```
VITE_API_BASE_URL=http://localhost:3001
```
然后重新构建。

### **Q5: 构建后文件很大？**

Vite 会自动进行代码分割和压缩。可以运行以下命令分析包大小：
```bash
npm install -D rollup-plugin-visualizer
# 在 vite.config.mjs 中添加 visualizer 插件
```

---

## 📚 相关命令总结

| 命令 | 用途 |
|------|------|
| `npm run dev:all` | 启动本地开发（Mock 后端 + Vite） |
| `npm run build` | 生产构建 |
| `npm run preview` | 本地预览生产构建 |
| `npm install` | 安装依赖 |

---

## 🎯 部署流程图

```
开发阶段
├─ npm install              # 安装依赖
├─ npm run dev:all          # 启动 Mock + 前端开发
└─ 测试功能                 # 在 http://localhost:5173

生产部署
├─ npm run build            # 构建（使用 .env.production）
├─ 上传 dist/ 到服务器      # scp 或 FTP
├─ 配置 Web 服务器          # Nginx / Apache
├─ 配置 API 代理            # 指向真实后端
└─ 访问部署地址             # http://your-server

上线运维
├─ 监控构建输出             # 检查错误和警告
├─ 测试所有功能             # 尤其是 API 集成
├─ 检查浏览器兼容性         # 跨浏览器测试
└─ 性能优化                 # 分析包大小、加载时间
```

---

## 📞 需要帮助？

如果部署过程中遇到问题：
1. 检查 `.env.production` 中的服务器地址是否正确
2. 在浏览器 DevTools 查看 Network 请求
3. 查看服务器日志，确认后端是否接收请求
4. 确认防火墙和网络配置允许前端访问后端

---

**部署愉快！** 🚀
