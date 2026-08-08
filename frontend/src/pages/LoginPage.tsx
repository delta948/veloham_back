import { useState } from 'react';
import axios from 'axios';
import { Link, useNavigate } from 'react-router-dom';
import { useAuthStore } from '../store/auth';

export function LoginPage() {
  const [loginValue, setLoginValue] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const login = useAuthStore((s) => s.login);
  const blockedReason = useAuthStore((s) => s.blockedReason);
  const navigate = useNavigate();

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
	try {
		await login(loginValue, password);
		navigate('/profile');
	} catch (requestError) {
		if (axios.isAxiosError(requestError) && !requestError.response) setError('Сервер недоступен. Попробуйте ещё раз.');
		else if (axios.isAxiosError(requestError) && requestError.response?.status === 403 && requestError.response.data?.error === 'account_blocked') setError('');
		else setError('Неверный логин или пароль');
	}
  };

  return (
    <form onSubmit={submit} className="panel mx-auto max-w-md space-y-4 p-4 sm:p-6">
      <h1 className="text-2xl font-black uppercase sm:text-4xl">Вход</h1>
      {blockedReason && <div className="border border-danger bg-danger/15 p-4 text-left" role="alert"><b className="block text-xl text-danger">Ваш аккаунт заблокирован</b><p className="mt-2 text-white/80">Причина: {blockedReason}</p></div>}
      {error && <div className="bg-danger p-3 font-bold">{error}</div>}
      <input className="field" required autoComplete="username" placeholder="Email или логин" value={loginValue} onChange={(e) => setLoginValue(e.target.value)} />
      <input className="field" type="password" placeholder="Пароль" value={password} onChange={(e) => setPassword(e.target.value)} />
      <button className="btn w-full">Войти</button>
      <Link className="block text-center text-white/65" to="/forgot-password">Забыли пароль?</Link>
      <Link className="block text-center text-acid" to="/register">Создать аккаунт</Link>
    </form>
  );
}
