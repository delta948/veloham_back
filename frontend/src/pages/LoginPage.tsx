import { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useAuthStore } from '../store/auth';

export function LoginPage() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const login = useAuthStore((s) => s.login);
  const navigate = useNavigate();

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    try { await login(email, password); navigate('/profile'); } catch { setError('Неверный email или пароль'); }
  };

  return (
    <form onSubmit={submit} className="panel mx-auto max-w-md space-y-4 p-6">
      <h1 className="text-4xl font-black uppercase">Вход</h1>
      {error && <div className="bg-danger p-3 font-bold">{error}</div>}
      <input className="field" placeholder="Email" value={email} onChange={(e) => setEmail(e.target.value)} />
      <input className="field" type="password" placeholder="Пароль" value={password} onChange={(e) => setPassword(e.target.value)} />
      <button className="btn w-full">Войти</button>
      <Link className="block text-center text-acid" to="/register">Создать аккаунт</Link>
    </form>
  );
}
