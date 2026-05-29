import axios from 'axios';
import { authStore } from '../stores/auth';

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
    const isAuthRequest = requestUrl.includes('/auth/login') || requestUrl.includes('/auth/register');

    if (status === 401 && !isAuthRequest) {
      authStore.clearAuth();
      if (!window.location.pathname.startsWith('/login')) {
        const redirect = encodeURIComponent(window.location.pathname + window.location.search);
        window.location.href = `/login?redirect=${redirect}`;
      }
    }

    return Promise.reject(error);
  }
);
