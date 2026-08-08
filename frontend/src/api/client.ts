import axios from 'axios';
import { useAuthStore } from '../store/auth';

const browserOrigin = typeof window === 'undefined' ? '' : window.location.origin;
const browserHost = typeof window === 'undefined' ? '127.0.0.1' : window.location.hostname;

export const API_BASE = import.meta.env.VITE_API_URL ?? (import.meta.env.DEV ? `http://${browserHost}:8080/api/v1` : '/api/v1');
export const WS_BASE = import.meta.env.VITE_WS_URL ?? (import.meta.env.DEV ? `ws://${browserHost}:8080/ws` : `${browserOrigin.replace(/^http/, 'ws')}/ws`);
export const MEDIA_BASE = API_BASE.replace(/\/api(?:\/v1)?\/?$/, '');

export const api = axios.create({ baseURL: API_BASE });

api.interceptors.request.use((config) => {
  const token = useAuthStore.getState().token;
  if (token) config.headers.Authorization = `Bearer ${token}`;
  return config;
});
