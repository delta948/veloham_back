import { Heart, MapPin, Wrench } from 'lucide-react';
import { Link } from 'react-router-dom';
import { api } from '../api/client';
import { labelTone } from '../constants/listingMeta';
import type { Listing } from '../types';
import { formatSom } from '../utils/format';
import { listingImageUrl } from '../utils/media';
import { bikeTypeLabel } from '../utils/sizeFit';
import { PriceDropBadge } from './PriceDropBadge';

export function ListingCard({ listing, onFavorite }: { listing: Listing; onFavorite?: () => void }) {
  const image = listingImageUrl(listing.images?.[0]?.image_url);
  const addFavorite = async () => {
    await api.post(`/favorites/${listing.id}`);
    onFavorite?.();
  };

  return (
    <article className="group panel overflow-hidden transition hover:-translate-y-1 hover:shadow-street">
      <div className="aspect-[4/3] overflow-hidden bg-black">
        <img className="h-full w-full object-cover transition duration-500 group-hover:scale-110" src={image} alt={listing.title} />
      </div>
      <div className="space-y-4 p-4">
        <div className="flex items-start justify-between gap-3">
          <h3 className="line-clamp-2 text-xl font-black uppercase">{listing.title}</h3>
          <button title="В избранное" onClick={addFavorite} className="border border-white/20 p-2 text-acid transition hover:bg-acid hover:text-black">
            <Heart size={18} />
          </button>
        </div>
        <PriceDropBadge listing={listing} detailed />
        <div className="text-3xl font-black text-acid">{formatSom(listing.price)}</div>
        <div className="grid grid-cols-2 gap-2 text-sm text-white/70">
          <span className="flex items-center gap-1"><MapPin size={15} /> {listing.city}</span>
          <span className="flex items-center gap-1"><Wrench size={15} /> {listing.condition}</span>
          {listing.brand && <span className="bg-black px-2 py-1 font-bold uppercase text-white/70">{listing.brand}</span>}
          {listing.bike_type && <span className="bg-black px-2 py-1 font-bold uppercase text-white">{bikeTypeLabel(listing.bike_type)}</span>}
          {listing.frame_size && <span className="bg-black px-2 py-1 font-bold uppercase text-acid">{listing.frame_size}</span>}
          <span className="bg-black px-2 py-1 font-bold uppercase text-white">{listing.category}</span>
          <span className="bg-black px-2 py-1 font-bold uppercase text-danger">{listing.status}</span>
        </div>
        <div className="flex flex-wrap gap-2 text-xs font-black uppercase">
          <span className="border border-white/15 bg-black px-2 py-1 text-white/70">{listing.deal_type ?? 'продажа'}</span>
          {(listing.labels ?? []).map((label) => (
            <span key={label} className={`px-2 py-1 ${labelTone(label)}`}>{label}</span>
          ))}
          {listing.match_preference?.exchange_enabled && <span className="bg-acid px-2 py-1 text-black">Обмен доступен</span>}
        </div>
        <Link className="btn w-full" to={`/listing/${listing.id}`}>Подробнее</Link>
      </div>
    </article>
  );
}
