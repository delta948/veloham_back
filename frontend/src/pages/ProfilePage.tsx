import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { api } from '../api/client';
import { ListingCard } from '../components/ListingCard';
import { useAuthStore } from '../store/auth';
import type { Chat, Favorite, Listing } from '../types';

export function ProfilePage() {
  const user = useAuthStore((s) => s.user);
  const loadMe = useAuthStore((s) => s.loadMe);
  const [listings, setListings] = useState<Listing[]>([]);
  const [favorites, setFavorites] = useState<Favorite[]>([]);
  const [chats, setChats] = useState<Chat[]>([]);
  const [profile, setProfile] = useState({ username: '', city: '', contact: '', avatar_url: '', password: '' });
  useEffect(() => { api.get<Listing[]>('/listings').then(({ data }) => setListings(data.filter((x) => x.user_id === user?.id))); }, [user?.id]);
  useEffect(() => {
    if (!user) return;
    setProfile({ username: user.username, city: user.city ?? '', contact: user.contact ?? '', avatar_url: user.avatar_url ?? '', password: '' });
    api.get<Favorite[]>('/favorites').then(({ data }) => setFavorites(data));
    api.get<Chat[]>('/chats').then(({ data }) => setChats(data));
  }, [user]);

  const save = async (e: React.FormEvent) => {
    e.preventDefault();
    await api.put('/users/me', profile);
    await loadMe();
  };

  return (
    <div className="space-y-8">
      <section className="panel flex flex-col justify-between gap-5 p-6 md:flex-row md:items-center">
        <div>
          <h1 className="text-5xl font-black uppercase">{user?.username}</h1>
          <p className="text-white/60">{user?.email} · {user?.city || 'город не указан'} · рейтинг {Number(user?.rating ?? 0).toFixed(1)}</p>
          <p className="text-white/50">Избранное: {favorites.length} · Чаты: {chats.length} · Сделки: история появится после отзывов</p>
        </div>
        <Link className="btn" to="/create">Новое объявление</Link>
      </section>
      <form onSubmit={save} className="panel grid gap-3 p-5 md:grid-cols-5">
        <input className="field" placeholder="Имя" value={profile.username} onChange={(e) => setProfile({ ...profile, username: e.target.value })} />
        <input className="field" placeholder="Город" value={profile.city} onChange={(e) => setProfile({ ...profile, city: e.target.value })} />
        <input className="field" placeholder="Телефон или Telegram" value={profile.contact} onChange={(e) => setProfile({ ...profile, contact: e.target.value })} />
        <input className="field" placeholder="URL аватара" value={profile.avatar_url} onChange={(e) => setProfile({ ...profile, avatar_url: e.target.value })} />
        <input className="field" type="password" placeholder="Новый пароль" value={profile.password} onChange={(e) => setProfile({ ...profile, password: e.target.value })} />
        <button className="btn md:col-span-5">Сохранить профиль</button>
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
