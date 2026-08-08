import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { api } from '../api/client';
import { ListingCard } from '../components/ListingCard';
import { useAuthStore } from '../store/auth';
import type { Chat, Favorite, Listing } from '../types';
import { avatarImageUrl } from '../utils/media';

export function ProfilePage() {
  const user = useAuthStore((s) => s.user);
  const loadMe = useAuthStore((s) => s.loadMe);
  const [listings, setListings] = useState<Listing[]>([]);
  const [favorites, setFavorites] = useState<Favorite[]>([]);
  const [chats, setChats] = useState<Chat[]>([]);
  const [profile, setProfile] = useState({ username: '', city: '', contact: '', password: '' });
  const [avatar, setAvatar] = useState<File | null>(null);
  const [avatarPreview, setAvatarPreview] = useState('');
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState('');
  useEffect(() => { api.get<Listing[]>('/listings').then(({ data }) => setListings(data.filter((x) => x.user_id === user?.id))); }, [user?.id]);
  useEffect(() => {
    if (!user) return;
    setProfile({ username: user.username, city: user.city ?? '', contact: user.contact ?? '', password: '' });
    api.get<Favorite[]>('/favorites').then(({ data }) => setFavorites(data));
    api.get<Chat[]>('/chats').then(({ data }) => setChats(data));
  }, [user]);

  const save = async (e: React.FormEvent) => {
    e.preventDefault();
    setSaving(true);
    setMessage('');
    try {
      await api.put('/users/me', profile);
      if (avatar) {
        const form = new FormData();
        form.append('avatar', avatar);
        await api.post('/users/me/avatar', form);
      }
      await loadMe();
      setAvatar(null);
      setAvatarPreview('');
      setProfile((current) => ({ ...current, password: '' }));
      setMessage('Профиль сохранён');
    } catch {
      setMessage('Не удалось сохранить профиль. Фото должно быть JPEG, PNG или WebP до 5 МБ.');
    } finally {
      setSaving(false);
    }
  };

  const selectAvatar = (file?: File) => {
    if (!file) return;
    if (avatarPreview) URL.revokeObjectURL(avatarPreview);
    setAvatar(file);
    setAvatarPreview(URL.createObjectURL(file));
  };

  return (
    <div className="space-y-6 sm:space-y-8">
      <section className="panel flex flex-col justify-between gap-5 p-4 sm:p-6 md:flex-row md:items-center">
        <div className="min-w-0">
          <h1 className="break-words text-3xl font-black uppercase sm:text-5xl">{user?.username}</h1>
          <p className="break-words text-sm text-white/60 sm:text-base">{user?.phone || user?.email} · {user?.city || 'город не указан'} · рейтинг {Number(user?.rating ?? 0).toFixed(1)}</p>
          <p className="text-white/50">Избранное: {favorites.length} · Чаты: {chats.length} · Сделки: история появится после отзывов</p>
          <div className="mt-3 flex flex-wrap gap-3 text-sm font-black uppercase"><Link className="text-acid" to="/favorites">Открыть избранное</Link><Link className="text-acid" to="/chats">Открыть чаты</Link></div>
        </div>
        <Link className="btn w-full md:w-auto" to="/create">Новое объявление</Link>
      </section>
      <form onSubmit={save} className="panel grid gap-3 p-4 sm:p-5 md:grid-cols-5">
		<label className="flex cursor-pointer items-center gap-3 rounded-xl border border-white/15 p-3 md:col-span-5">
			<div className="avatar h-16 w-16 text-xl">{avatarPreview || user?.avatar_url ? <img src={avatarPreview || avatarImageUrl(user?.avatar_url)} alt="Аватар" /> : user?.username?.[0]}</div>
			<div><b className="block">Выбрать фото из галереи</b><span className="text-xs text-white/50">JPEG, PNG или WebP, до 5 МБ</span></div>
			<input className="sr-only" type="file" accept="image/jpeg,image/png,image/webp" onChange={(e) => selectAvatar(e.target.files?.[0])} />
		</label>
        <input className="field" placeholder="Имя" value={profile.username} onChange={(e) => setProfile({ ...profile, username: e.target.value })} />
        <input className="field" placeholder="Город" value={profile.city} onChange={(e) => setProfile({ ...profile, city: e.target.value })} />
        <input className="field" placeholder="Телефон или Telegram" value={profile.contact} onChange={(e) => setProfile({ ...profile, contact: e.target.value })} />
		<input className="field" type="password" minLength={6} maxLength={72} placeholder="Новый пароль (минимум 6 символов)" value={profile.password} onChange={(e) => setProfile({ ...profile, password: e.target.value })} />
		{message && <p className="md:col-span-5">{message}</p>}
		<button className="btn md:col-span-5" disabled={saving}>{saving ? 'Сохраняем…' : 'Сохранить профиль'}</button>
      </form>
      <section className="space-y-5">
        <h2 className="text-3xl font-black uppercase">Мои объявления</h2>
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
          {listings.map((listing) => <ListingCard key={listing.id} listing={listing} />)}
        </div>
      </section>
    </div>
  );
}
