import { defineConfig } from 'vite';
import vue from '@vitejs/plugin-vue';

export default defineConfig({
  plugins: [vue()],
  server: {
    // 开发环境：代理 /api/v1 请求到本地 Mock 服务器 (localhost:3001)
    // 生产环境：直接请求真实后端服务器（在 http.js 中配置）
    proxy: {
      '/api/v1': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
});

