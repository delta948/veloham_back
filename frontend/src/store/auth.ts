import { create } from 'zustand';
import { api } from '../api/client';
import type { User } from '../types';

type AuthState = {
  token: string | null;
  user: User | null;
  login: (email: string, password: string) => Promise<void>;
  register: (username: string, email: string, password: string) => Promise<void>;
  loadMe: () => Promise<void>;
  logout: () => void;
};

const savedToken = localStorage.getItem('veloham_token');

export const useAuthStore = create<AuthState>((set) => ({
  token: savedToken,
  user: null,
  async login(email, password) {
    const { data } = await api.post('/auth/login', { email, password });
    localStorage.setItem('veloham_token', data.token);
    set({ token: data.token, user: data.user });
  },
  async register(username, email, password) {
    const { data } = await api.post('/auth/register', { username, email, password });
    localStorage.setItem('veloham_token', data.token);
    set({ token: data.token, user: data.user });
  },
  async loadMe() {
    if (!useAuthStore.getState().token) return;
    const { data } = await api.get('/auth/me');
    set({ user: data });
  },
  logout() {
    localStorage.removeItem('veloham_token');
    set({ token: null, user: null });
  }
}));
