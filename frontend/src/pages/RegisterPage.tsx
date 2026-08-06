import { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useAuthStore } from '../store/auth';

export function RegisterPage() {
  const [username, setUsername] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const register = useAuthStore((s) => s.register);
  const navigate = useNavigate();

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    await register(username, email, password);
    navigate('/profile');
  };

  return (
    <form onSubmit={submit} className="panel mx-auto max-w-md space-y-4 p-6">
      <h1 className="text-4xl font-black uppercase">Регистрация</h1>
      <input className="field" placeholder="Имя" value={username} onChange={(e) => setUsername(e.target.value)} />
      <input className="field" placeholder="Email" value={email} onChange={(e) => setEmail(e.target.value)} />
      <input className="field" type="password" minLength={12} maxLength={72} placeholder="Пароль (минимум 12 символов)" value={password} onChange={(e) => setPassword(e.target.value)} />
      <button className="btn w-full">Зарегистрироваться</button>
      <Link className="block text-center text-acid" to="/login">Уже есть аккаунт</Link>
    </form>
  );
}
