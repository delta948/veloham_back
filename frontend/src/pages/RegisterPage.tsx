import { useState } from 'react';
import axios from 'axios';
import { Link, useNavigate } from 'react-router-dom';
import { useAuthStore } from '../store/auth';
import type { RegistrationStart } from '../store/auth';

export function RegisterPage() {
  const [username, setUsername] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [code, setCode] = useState('');
  const [verification, setVerification] = useState<RegistrationStart | null>(null);
  const [error, setError] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const register = useAuthStore((s) => s.register);
  const verifyRegistration = useAuthStore((s) => s.verifyRegistration);
  const resendRegistration = useAuthStore((s) => s.resendRegistration);
  const navigate = useNavigate();

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setSubmitting(true);
    try {
      const result = await register(username.trim(), email.trim(), password);
      setVerification(result);
    } catch (requestError) {
      if (axios.isAxiosError(requestError) && !requestError.response) {
        setError('Сервер недоступен. Попробуйте ещё раз.');
      } else if (axios.isAxiosError(requestError) && requestError.response?.status === 409) {
        setError('Этот email уже зарегистрирован.');
      } else if (axios.isAxiosError(requestError) && requestError.response?.status === 429) {
        setError('Слишком много попыток. Подождите минуту.');
      } else if (axios.isAxiosError(requestError) && requestError.response?.status === 503) {
        setError('Отправка писем пока не настроена. Обратитесь к администратору.');
      } else {
        setError('Проверьте имя, email и пароль.');
      }
    } finally {
      setSubmitting(false);
    }
  };

  const verify = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!verification) return;
    setError('');
    setSubmitting(true);
    try {
      await verifyRegistration(verification.verification_id, code.trim());
      navigate('/profile');
    } catch (requestError) {
      if (axios.isAxiosError(requestError) && !requestError.response) setError('Сервер недоступен. Попробуйте ещё раз.');
      else if (axios.isAxiosError(requestError) && requestError.response?.status === 429) setError('Слишком много неверных попыток. Запросите новый код.');
      else setError('Неверный код или срок его действия истёк.');
    } finally {
      setSubmitting(false);
    }
  };

  const resend = async () => {
    if (!verification) return;
    setError('');
    try {
      setVerification(await resendRegistration(verification.verification_id));
      setCode('');
    } catch (requestError) {
      if (axios.isAxiosError(requestError) && requestError.response?.status === 429) setError('Повторный код можно запросить через минуту.');
      else if (axios.isAxiosError(requestError) && requestError.response?.status === 503) setError('Отправка писем пока не настроена.');
      else setError('Не удалось отправить новый код.');
    }
  };

  if (verification) {
    return (
      <form onSubmit={verify} className="panel mx-auto max-w-md space-y-4 p-4 sm:p-6">
        <h1 className="break-words text-2xl font-black uppercase sm:text-4xl">Подтвердите номер</h1>
        <p className="text-white/65">Код отправлен на почту <b className="text-white">{verification.email}</b>. Код действует 10 минут.</p>
        {error && <div className="bg-danger p-3 font-bold" role="alert">{error}</div>}
        <input className="field text-center text-2xl tracking-[.3em]" required autoFocus inputMode="numeric" autoComplete="one-time-code" minLength={4} maxLength={12} placeholder="Код из письма" value={code} onChange={(e) => setCode(e.target.value)} />
        <button className="btn w-full" disabled={submitting}>{submitting ? 'Проверяем…' : 'Подтвердить и зарегистрироваться'}</button>
        <button className="btn-dark w-full" type="button" onClick={() => void resend()}>Отправить код повторно</button>
        <button className="block w-full text-center text-white/55" type="button" onClick={() => { setVerification(null); setCode(''); setError(''); }}>Изменить данные</button>
      </form>
    );
  }

  return (
    <form onSubmit={submit} className="panel mx-auto max-w-md space-y-4 p-4 sm:p-6">
      <h1 className="break-words text-2xl font-black uppercase sm:text-4xl">Регистрация</h1>
      {error && <div className="bg-danger p-3 font-bold" role="alert">{error}</div>}
      <input className="field" required autoComplete="name" placeholder="Имя" value={username} onChange={(e) => setUsername(e.target.value)} />
		<input className="field" required type="email" autoComplete="email" inputMode="email" placeholder="Email" value={email} onChange={(e) => setEmail(e.target.value)} />
		<input className="field" required type="password" autoComplete="new-password" minLength={6} maxLength={72} placeholder="Пароль (минимум 6 символов)" value={password} onChange={(e) => setPassword(e.target.value)} />
      <button className="btn w-full" disabled={submitting}>{submitting ? 'Отправляем письмо…' : 'Получить код на почту'}</button>
      <Link className="block text-center text-acid" to="/login">Уже есть аккаунт</Link>
    </form>
  );
}
