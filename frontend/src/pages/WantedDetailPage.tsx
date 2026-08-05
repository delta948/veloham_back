import { MessageSquare } from 'lucide-react';
import { useEffect, useState } from 'react';
import { Link, useNavigate, useParams } from 'react-router-dom';
import { api } from '../api/client';
import { ListingCard } from '../components/ListingCard';
import { useAuthStore } from '../store/auth';
import type { Listing, WantedOffer, WantedRequest } from '../types';
import { formatSom } from '../utils/format';

type WantedResponse = {
  request: WantedRequest;
  matches: Listing[];
};

export function WantedDetailPage() {
  const { id } = useParams();
  const navigate = useNavigate();
  const user = useAuthStore((s) => s.user);
  const token = useAuthStore((s) => s.token);
  const [request, setRequest] = useState<WantedRequest | null>(null);
  const [matches, setMatches] = useState<Listing[]>([]);
  const [myListings, setMyListings] = useState<Listing[]>([]);
  const [listingId, setListingId] = useState('');
  const [message, setMessage] = useState('');

  const load = () => {
    if (!id) return;
    api.get<WantedResponse>(`/wanted/${id}`).then(({ data }) => {
      setRequest(data.request);
      setMatches(data.matches);
    });
  };

  useEffect(() => {
    load();
  }, [id]);

  useEffect(() => {
    if (!user?.id) return;
    api.get<Listing[]>(`/users/${user.id}/listings`).then(({ data }) => {
      setMyListings(data);
      setListingId(data[0]?.id ?? '');
    });
  }, [user?.id]);

  if (!request) return <div className="text-white/60">Загрузка...</div>;

  const offer = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!token) {
      navigate('/login');
      return;
    }
    await api.post(`/wanted/${request.id}/offers`, { listing_id: listingId, message });
    setMessage('');
    load();
  };

  const close = async () => {
    await api.patch(`/wanted/${request.id}/close`);
    load();
  };

  const startChat = async (listing: Listing) => {
    const { data } = await api.post('/chats', { listing_id: listing.id });
    navigate(`/chats/${data.id}`);
  };

  return (
    <div className="space-y-8">
      <section className="panel p-6">
        <div className="flex flex-col justify-between gap-4 md:flex-row md:items-start">
          <div>
            <div className="flex flex-wrap gap-2">
              <span className="bg-acid px-3 py-1 font-black uppercase text-black">{request.category}</span>
              <span className="bg-black px-3 py-1 font-black uppercase text-white/70">{request.status === 'closed' ? 'Закрыто' : 'Активно'}</span>
            </div>
            <h1 className="mt-4 text-5xl font-black uppercase">{request.title}</h1>
            <p className="mt-3 text-white/65">{request.description}</p>
            <p className="mt-4 text-white/75">{request.city} · ростовка {request.frame_size || 'любая'} · рост {request.rider_height || 'не указан'} см</p>
            <div className="mt-3 text-3xl font-black text-acid">{formatSom(request.min_budget)} - {formatSom(request.max_budget)}</div>
          </div>
          {user?.id === request.user_id && request.status !== 'closed' && <button className="btn-dark" onClick={close}>Отметить закрытой</button>}
        </div>
      </section>

      {user?.id !== request.user_id && (
        <form onSubmit={offer} className="panel grid gap-3 p-5 md:grid-cols-[1fr_2fr_auto]">
          <select className="field" value={listingId} onChange={(e) => setListingId(e.target.value)}>
            {myListings.map((listing) => <option key={listing.id} value={listing.id}>{listing.title}</option>)}
          </select>
          <input className="field" placeholder="Сообщение покупателю" value={message} onChange={(e) => setMessage(e.target.value)} />
          <button className="btn">Предложить товар</button>
        </form>
      )}

      <section className="space-y-5">
        <h2 className="text-3xl font-black uppercase">Подходящие объявления</h2>
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
          {matches.map((listing) => <ListingCard key={listing.id} listing={listing} />)}
        </div>
      </section>

      <section className="space-y-5">
        <h2 className="text-3xl font-black uppercase">Предложения продавцов</h2>
        <div className="grid gap-4">
          {(request.offers ?? []).map((offer: WantedOffer) => (
            <div key={offer.id} className="panel flex flex-col justify-between gap-4 p-4 md:flex-row md:items-center">
              <div>
                <h3 className="text-xl font-black uppercase">{offer.listing?.title}</h3>
                <p className="text-white/60">{offer.seller?.username}: {offer.message || 'Предлагаю свой товар'}</p>
              </div>
              <div className="flex flex-wrap gap-2">
                <Link className="btn-dark px-3 py-2" to={`/listing/${offer.listing_id}`}>Открыть объявление</Link>
                <button className="btn px-3 py-2" onClick={() => startChat(offer.listing)}><MessageSquare size={16} /> Чат</button>
              </div>
            </div>
          ))}
          {(request.offers ?? []).length === 0 && <p className="text-white/50">Предложений пока нет.</p>}
        </div>
      </section>
    </div>
  );
}
