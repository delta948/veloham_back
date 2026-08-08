import { create } from 'zustand';
import axios from 'axios';
import { api } from '../api/client';
import type { User } from '../types';

type AuthState = {
  token: string | null;
  user: User | null;
  blockedReason: string;
  login: (login: string, password: string) => Promise<void>;
  register: (username: string, email: string, password: string) => Promise<RegistrationStart>;
  verifyRegistration: (verificationId: string, code: string) => Promise<void>;
  resendRegistration: (verificationId: string) => Promise<RegistrationStart>;
  loadMe: () => Promise<void>;
  logout: () => void;
};

export type RegistrationStart = {
  verification_id: string;
  email: string;
  expires_in: number;
  dev_code?: string;
};

const savedToken = localStorage.getItem('veloham_token');

export const useAuthStore = create<AuthState>((set) => ({
  token: savedToken,
  user: null,
  blockedReason: '',
  async login(login, password) {
    try {
      const { data } = await api.post('/auth/login', { login, password });
      localStorage.setItem('veloham_token', data.token);
      set({ token: data.token, user: data.user, blockedReason: '' });
    } catch (error) {
      if (axios.isAxiosError(error) && error.response?.status === 403 && error.response.data?.error === 'account_blocked') {
        localStorage.removeItem('veloham_token');
        set({ token: null, user: null, blockedReason: error.response.data.reason || 'Причина не указана' });
      }
      throw error;
    }
  },
  async register(username, email, password) {
    const { data } = await api.post<RegistrationStart>('/auth/register', { username, email, password });
    return data;
  },
  async verifyRegistration(verificationId, code) {
    const { data } = await api.post('/auth/register/verify', { verification_id: verificationId, code });
    localStorage.setItem('veloham_token', data.token);
    set({ token: data.token, user: data.user, blockedReason: '' });
  },
  async resendRegistration(verificationId) {
    const { data } = await api.post<RegistrationStart>('/auth/register/resend', { verification_id: verificationId });
    return data;
  },
  async loadMe() {
    if (!useAuthStore.getState().token) return;
    try {
      const { data } = await api.get('/auth/me');
      set({ user: data, blockedReason: '' });
    } catch (error) {
      if (axios.isAxiosError(error) && error.response?.status === 403 && error.response.data?.error === 'account_blocked') {
        localStorage.removeItem('veloham_token');
        set({ token: null, user: null, blockedReason: error.response.data.reason || 'Причина не указана' });
        return;
      }
      throw error;
    }
  },
  logout() {
    localStorage.removeItem('veloham_token');
    set({ token: null, user: null, blockedReason: '' });
  }
}));
