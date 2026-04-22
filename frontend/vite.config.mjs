import { defineConfig } from 'vite';
import vue from '@vitejs/plugin-vue';

const devApiTarget = process.env.VITE_DEV_API_TARGET || 'http://127.0.0.1:8080';

export default defineConfig({
  plugins: [vue()],
  server: {
    // 开发环境：代理 /api/v1 请求到 Go 后端。
    // 如需切换目标，可设置环境变量 VITE_DEV_API_TARGET。
    proxy: {
      '/api/v1': {
        //只有Mock测试时使用localhost的3001端口，其余时间选择target: devApiTarget
        target: devApiTarget,
        //target: 'http://127.0.0.1:3001',
        changeOrigin: true,
      },
    },
  },
});

