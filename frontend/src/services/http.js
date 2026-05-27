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
  withCredentials: true,
  headers: {
    'Content-Type': 'application/json',
  },
});

http.interceptors.response.use(
  (response) => response,
  (error) => {
    const status = error?.response?.status;
    const requestUrl = String(error?.config?.url || '');

    // Ignore auth pages' own 401 handling.
    const isAuthRequest = requestUrl.includes('/auth/login') || requestUrl.includes('/auth/register');
    
    if (status === 401 && !isAuthRequest) {
      authStore.clearAuth();
      if (!window.location.pathname.startsWith('/login')) {
        const redirect = encodeURIComponent(window.location.pathname + window.location.search);
        window.location.href = `/login?redirect=${redirect}`;
      }
    }

    // 403 越权操作拦截
    if (status === 403) {
      alert('您的操作权限不足，系统已重置您的状态。');
      authStore.clearAuth(); // 清理前端状态
      window.location.href = '/login'; // 安全切回登录页，彻底切断原页面的异步重试死循环
}

    return Promise.reject(error);
  }
);