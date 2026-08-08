import { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import { api } from '../api/client';
import { avatarImageUrl } from '../utils/media';
import { ListingCard } from '../components/ListingCard';
import type { Listing, Review, User } from '../types';

export function SellerProfilePage() {
  const { id } = useParams();
  const [seller, setSeller] = useState<User | null>(null);
  const [listings, setListings] = useState<Listing[]>([]);
  const [reviews, setReviews] = useState<Review[]>([]);

  useEffect(() => {
    if (!id) return;
    api.get<User>(`/users/${id}`).then(({ data }) => setSeller(data));
    api.get<Listing[]>(`/users/${id}/listings`).then(({ data }) => setListings(data));
    api.get<Review[]>(`/users/${id}/reviews`).then(({ data }) => setReviews(data));
  }, [id]);

  if (!seller) return <div className="text-white/60">Загрузка...</div>;

  return (
    <div className="space-y-6 sm:space-y-8">
      <section className="panel flex flex-col gap-5 p-4 sm:p-6 md:flex-row md:items-center">
		<img className="h-24 w-24 border border-white/15 object-cover" src={avatarImageUrl(seller.avatar_url) || 'https://images.unsplash.com/photo-1508214751196-bcfd4ca60f91?q=80&w=400&auto=format&fit=crop'} alt={seller.username} />
        <div className="min-w-0">
          <h1 className="break-words text-3xl font-black uppercase sm:text-5xl">{seller.username}</h1>
          <p className="break-words text-sm text-white/60 sm:text-base">{seller.city || 'Кыргызстан'} · рейтинг {Number(seller.rating ?? 0).toFixed(1)} · с {new Date(seller.created_at).toLocaleDateString('ru-KG')}</p>
          <p className="text-white/50">Активных объявлений: {listings.length}</p>
        </div>
      </section>
      <section className="space-y-4">
        <h2 className="text-3xl font-black uppercase">Отзывы</h2>
        <div className="grid gap-3">
          {reviews.map((review) => (
            <div key={review.id} className="panel p-4">
              <div className="font-black text-acid">{review.rating}/5 · {review.author?.username}</div>
              <p className="text-white/70">{review.text}</p>
            </div>
          ))}
          {reviews.length === 0 && <p className="text-white/50">Отзывов пока нет.</p>}
        </div>
      </section>
      <section className="space-y-4">
        <h2 className="text-3xl font-black uppercase">Объявления продавца</h2>
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
          {listings.map((listing) => <ListingCard key={listing.id} listing={listing} />)}
        </div>
      </section>
    </div>
  );
}
