import { http } from './http';

export const registerApi = (payload) => http.post('/auth/register', payload);

export const loginApi = (payload) => http.post('/auth/login', payload);

export const changePasswordApi = (payload) => http.put('/auth/password', payload);
