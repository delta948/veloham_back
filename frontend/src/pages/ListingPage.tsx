import { AlertTriangle, Heart, MessageSquare, Pencil, Trash2 } from 'lucide-react';
import { useEffect, useState } from 'react';
import { Link, useNavigate, useParams } from 'react-router-dom';
import { api } from '../api/client';
import { getListing } from '../api/listings';
import { ListingCard } from '../components/ListingCard';
import { labelTone } from '../constants/listingMeta';
import { useAuthStore } from '../store/auth';
import type { Listing } from '../types';
import { formatSom } from '../utils/format';
import { listingImageUrl } from '../utils/media';
import { bikeTypeLabel } from '../utils/sizeFit';
import { PriceDropBadge } from '../components/PriceDropBadge';
import { PriceHistory } from '../components/PriceHistory';

export function ListingPage() {
  const { id } = useParams();
  const [listing, setListing] = useState<Listing | null>(null);
  const [matches, setMatches] = useState<Listing[]>([]);
  const [similar, setSimilar] = useState<Listing[]>([]);
  const [reportStatus, setReportStatus] = useState('');
  const user = useAuthStore((s) => s.user);
  const token = useAuthStore((s) => s.token);
  const navigate = useNavigate();

  useEffect(() => {
    if (!id) return;
    getListing(id).then((item) => {
      setListing(item);
      api.get<Listing[]>(`/listings/${item.id}/matches`).then(({ data }) => setMatches(data)).catch(() => setMatches([]));
      api.get<Listing[]>('/listings', { params: { category: item.category } }).then(({ data }) => setSimilar(data.filter((x) => x.id !== item.id).slice(0, 3)));
    });
  }, [id]);
  if (!listing) return <div className="text-white/60">Загрузка...</div>;

  const image = listingImageUrl(listing.images?.[0]?.image_url);
  const startChat = async () => {
    const { data } = await api.post('/chats', { listing_id: listing.id });
    navigate(`/chats/${data.id}`);
  };
  const remove = async () => {
    await api.delete(`/listings/${listing.id}`);
    navigate('/market');
  };
  const report = async () => {
    if (!token) {
      navigate('/login');
      return;
    }
    setReportStatus('');
    try {
      await api.post('/reports', { listing_id: listing.id, seller_id: listing.user_id, reason: 'другое', text: 'Жалоба со страницы объявления' });
      setReportStatus('Жалоба отправлена администратору.');
    } catch {
      setReportStatus('Не удалось отправить жалобу. Попробуй еще раз.');
    }
  };
  const buildRows = listing.build_card ? Object.entries({
    'Рама': listing.build_card.frame,
    'Ростовка': listing.build_card.size,
    'Вилка': listing.build_card.fork,
    'Колеса': listing.build_card.wheels,
    'Втулки': listing.build_card.hubs,
    'Покрышки': listing.build_card.tires,
    'Руль': listing.build_card.handlebar,
    'Вынос': listing.build_card.stem,
    'Седло': listing.build_card.saddle,
    'Шатуны': listing.build_card.cranks,
    'Каретка': listing.build_card.bottom_bracket,
    'Цепь': listing.build_card.chain,
    'Cog / звезда': listing.build_card.cog,
    'Тормоза': listing.build_card.brakes,
    'Вес': listing.build_card.weight,
    'Состояние рамы': listing.build_card.frame_condition,
    'Дефекты': listing.build_card.defects,
    'Документы': listing.build_card.documents ? 'есть' : ''
  }).filter(([, value]) => value) : [];

  return (
    <div className="space-y-7 sm:space-y-10">
      <div className="grid gap-5 sm:gap-8 lg:grid-cols-[1.2fr_.8fr]">
        <div className="space-y-4">
          <img className="aspect-[16/10] w-full border border-white/10 object-cover shadow-danger" src={image} alt={listing.title} />
          <div className="hide-scroll flex snap-x gap-2 overflow-x-auto sm:grid sm:grid-cols-3 sm:gap-3">
            {listing.images?.slice(1).map((img) => <img key={img.id} className="aspect-video w-32 shrink-0 snap-start object-cover sm:w-full" src={listingImageUrl(img.image_url)} alt={listing.title} />)}
          </div>
        </div>
        <aside className="panel min-w-0 space-y-4 p-4 sm:space-y-5 sm:p-6">
          <div className="flex flex-wrap gap-2">
            <span className="bg-acid px-3 py-1 font-black uppercase text-black">{listing.category}</span>
            <span className="bg-danger px-3 py-1 font-black uppercase text-white">{listing.status}</span>
            <span className="border border-white/15 px-3 py-1 font-black uppercase">{listing.deal_type || 'продажа'}</span>
            {(listing.labels ?? []).map((label) => (
              <span key={label} className={`px-3 py-1 font-black uppercase ${labelTone(label)}`}>{label}</span>
            ))}
          </div>
          <h1 className="break-words text-3xl font-black uppercase sm:text-5xl">{listing.title}</h1>
          <PriceDropBadge listing={listing} detailed />
          <div className="break-words text-3xl font-black text-acid sm:text-5xl">{formatSom(listing.price)}</div>
          <p className="break-words text-white/70">{listing.description}</p>
          <div className="grid gap-2 text-white/80">
            <span>Город: <b>{listing.city}</b></span>
            {listing.brand && <span>Бренд: <b>{listing.brand}</b></span>}
            {listing.bike_type && <span>Тип велосипеда: <b>{bikeTypeLabel(listing.bike_type)}</b></span>}
            {listing.frame_size && <span>Ростовка: <b>{listing.frame_size}</b></span>}
            {(listing.rider_height_min || listing.rider_height_max) && <span>Рост райдера: <b>{listing.rider_height_min || '...'}-{listing.rider_height_max || '...'} см</b></span>}
            <span>Состояние: <b>{listing.condition}</b></span>
            <span>Дата: <b>{new Date(listing.created_at).toLocaleDateString('ru-KG')}</b></span>
            <Link to={`/profile/${listing.user_id}`} className="break-words text-acid">Продавец: <b>{listing.user?.username}</b> · рейтинг {Number(listing.user?.rating ?? 0).toFixed(1)}</Link>
          </div>
          <button className="btn w-full" onClick={startChat}><MessageSquare size={18} /> Написать продавцу</button>
          <button className="btn-dark w-full" onClick={() => api.post(`/favorites/${listing.id}`)}><Heart size={18} /> В избранное</button>
          <button className="btn-dark w-full" onClick={report}><AlertTriangle size={18} /> Пожаловаться</button>
          {reportStatus && <div className="border border-white/15 bg-black p-3 text-sm font-bold text-white/75">{reportStatus}</div>}
          {user?.id === listing.user_id && (
            <div className="grid gap-3 sm:grid-cols-2">
              <Link className="btn-dark" to={`/edit/${listing.id}`}><Pencil size={18} /> Редактировать</Link>
              <button className="btn-dark text-danger" onClick={remove}><Trash2 size={18} /> Удалить</button>
            </div>
          )}
        </aside>
      </div>

      <PriceHistory listingId={listing.id} createdAt={listing.created_at} />

      {buildRows.length > 0 && (
        <section className="panel p-4 sm:p-6">
          <h2 className="text-2xl font-black uppercase sm:text-4xl">Build Card</h2>
          <div className="mt-5 grid gap-3 md:grid-cols-3">
            {buildRows.map(([label, value]) => (
              <div key={label} className="border border-white/10 bg-black p-4">
                <div className="text-xs font-black uppercase text-acid">{label}</div>
                <div className="mt-1 font-bold text-white">{value}</div>
              </div>
            ))}
          </div>
        </section>
      )}

      {listing.match_preference?.exchange_enabled && (
        <section className="panel p-4 sm:p-6">
          <h2 className="text-2xl font-black uppercase sm:text-4xl">Bike Match</h2>
          <div className="mt-4 grid gap-3 text-white/75 md:grid-cols-3">
            <span>Хочет: <b>{listing.match_preference.wants || 'обмен'}</b></span>
            <span>Категории: <b>{listing.match_preference.categories || 'любые'}</b></span>
            <span>Диапазон: <b>{formatSom(listing.match_preference.min_price ?? 0)} - {formatSom(listing.match_preference.max_price ?? 0)}</b></span>
            <span>Доплата: <b>{listing.match_preference.can_add_cash ? `до ${formatSom(listing.match_preference.max_cash_add ?? 0)}` : 'нет'}</b></span>
            <span>Только свой город: <b>{listing.match_preference.same_city_only ? 'да' : 'нет'}</b></span>
          </div>
          {matches.length > 0 && (
            <div className="mt-6 grid gap-6 md:grid-cols-3">
              {matches.map((item) => <ListingCard key={item.id} listing={item} />)}
            </div>
          )}
        </section>
      )}

      {similar.length > 0 && (
        <section className="space-y-5">
          <h2 className="text-2xl font-black uppercase sm:text-4xl">Похожие объявления</h2>
          <div className="grid gap-6 md:grid-cols-3">
            {similar.map((item) => <ListingCard key={item.id} listing={item} />)}
          </div>
        </section>
      )}
    </div>
  );
}
