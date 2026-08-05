import axios from 'axios';
import { useAuthStore } from '../store/auth';

export const API_BASE = import.meta.env.VITE_API_URL ?? 'http://127.0.0.1:8080/api';
export const WS_BASE = import.meta.env.VITE_WS_URL ?? 'ws://127.0.0.1:8080/ws';
export const MEDIA_BASE = API_BASE.replace('/api', '');

export const api = axios.create({ baseURL: API_BASE });

api.interceptors.request.use((config) => {
  const token = useAuthStore.getState().token;
  if (token) config.headers.Authorization = `Bearer ${token}`;
  return config;
});
