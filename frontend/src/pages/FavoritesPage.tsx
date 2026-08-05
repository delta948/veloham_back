import { useEffect, useState } from 'react';
import { api } from '../api/client';
import { ListingCard } from '../components/ListingCard';
import type { Favorite } from '../types';

export function FavoritesPage() {
  const [favorites, setFavorites] = useState<Favorite[]>([]);
  const load = () => api.get<Favorite[]>('/favorites').then(({ data }) => setFavorites(data));
  useEffect(() => { void load(); }, []);

  return (
    <div className="space-y-6">
      <h1 className="text-5xl font-black uppercase">Избранное</h1>
      <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
        {favorites.map((favorite) => <ListingCard key={favorite.id} listing={favorite.listing} />)}
      </div>
    </div>
  );
}
