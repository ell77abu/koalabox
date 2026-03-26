import axios from 'axios';
import { authStore } from '../stores/auth';

/**
 * API 基础 URL 配置
 * - 开发环境：使用不需要 baseURL（因为 Vite 代理会处理）
 * - 生产环境：使用完整的服务器地址（例如 https://api.example.com）
 */
const isProduction = import.meta.env.PROD;
const baseURL = isProduction ? import.meta.env.VITE_API_BASE_URL : '/api/v1';

export const http = axios.create({
  baseURL,
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
});

http.interceptors.request.use((config) => {
  if (authStore.token.value) {
    config.headers.Authorization = `Bearer ${authStore.token.value}`;
  }
  return config;
});
