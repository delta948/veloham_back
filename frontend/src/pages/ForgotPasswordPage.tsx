import { useState } from 'react';
import axios from 'axios';
import { Link, useNavigate } from 'react-router-dom';
import { api } from '../api/client';

type ResetStart = { reset_id: string; expires_in: number };

export function ForgotPasswordPage() {
  const [email, setEmail] = useState('');
  const [reset, setReset] = useState<ResetStart | null>(null);
  const [code, setCode] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [busy, setBusy] = useState(false);
  const navigate = useNavigate();

  const requestCode = async (event: React.FormEvent) => {
    event.preventDefault(); setBusy(true); setError('');
    try {
      const { data } = await api.post<ResetStart>('/auth/password/forgot', { email: email.trim() });
      setReset(data);
    } catch (requestError) {
      if (axios.isAxiosError(requestError) && !requestError.response) setError('Сервер недоступен.');
      else if (axios.isAxiosError(requestError) && requestError.response?.status === 503) setError('Отправка писем пока не настроена.');
      else setError('Не удалось отправить письмо. Попробуйте позже.');
    } finally { setBusy(false); }
  };

  const changePassword = async (event: React.FormEvent) => {
    event.preventDefault(); if (!reset) return; setBusy(true); setError('');
    try {
      await api.post('/auth/password/reset', { reset_id: reset.reset_id, code: code.trim(), password });
      navigate('/login', { replace: true });
    } catch (requestError) {
      if (axios.isAxiosError(requestError) && requestError.response?.status === 429) setError('Слишком много попыток. Запросите новый код.');
      else setError('Неверный код или срок его действия истёк.');
    } finally { setBusy(false); }
  };

  if (reset) return (
    <form onSubmit={changePassword} className="panel mx-auto max-w-md space-y-4 p-4 sm:p-6">
      <h1 className="break-words text-2xl font-black uppercase sm:text-4xl">Новый пароль</h1>
      <p className="text-white/65">Если аккаунт существует, код уже отправлен на почту. Он действует 10 минут.</p>
      {error && <div className="bg-danger p-3 font-bold" role="alert">{error}</div>}
      <input className="field text-center text-2xl tracking-[.3em]" required inputMode="numeric" autoComplete="one-time-code" pattern="[0-9]{6}" maxLength={6} placeholder="Код из письма" value={code} onChange={(e) => setCode(e.target.value.replace(/\D/g, ''))} />
      <input className="field" required type="password" autoComplete="new-password" minLength={6} maxLength={72} placeholder="Новый пароль (минимум 6 символов)" value={password} onChange={(e) => setPassword(e.target.value)} />
      <button className="btn w-full" disabled={busy}>{busy ? 'Сохраняем…' : 'Сменить пароль'}</button>
      <button type="button" className="block w-full text-center text-white/55" onClick={() => { setReset(null); setCode(''); setError(''); }}>Запросить другой код</button>
    </form>
  );

  return (
    <form onSubmit={requestCode} className="panel mx-auto max-w-md space-y-4 p-4 sm:p-6">
      <h1 className="break-words text-2xl font-black uppercase sm:text-4xl">Восстановление</h1>
      <p className="text-white/65">Введите email аккаунта — отправим шестизначный код.</p>
      {error && <div className="bg-danger p-3 font-bold" role="alert">{error}</div>}
      <input className="field" required type="email" autoComplete="email" inputMode="email" placeholder="Email" value={email} onChange={(e) => setEmail(e.target.value)} />
      <button className="btn w-full" disabled={busy}>{busy ? 'Отправляем…' : 'Получить код'}</button>
      <Link className="block text-center text-acid" to="/login">Вернуться ко входу</Link>
    </form>
  );
}
