import { Search } from 'lucide-react';
import { useEffect, useState } from 'react';
import { Link, useSearchParams } from 'react-router-dom';
import { getListings, type ListingFilters } from '../api/listings';
import { ListingCard } from '../components/ListingCard';
import { CATEGORIES, categoryPath } from '../constants/catalog';
import { DEAL_TYPES, LISTING_LABELS, labelTone } from '../constants/listingMeta';
import type { Listing } from '../types';
import { BIKE_TYPES, recommendFrameSize, type BikeType } from '../utils/sizeFit';

export function MarketPage() {
  const [params] = useSearchParams();
  const [filters, setFilters] = useState<ListingFilters>({ category: params.get('category') ?? '' });
  const [listings, setListings] = useState<Listing[]>([]);
  const [fitHeight, setFitHeight] = useState('');
  const [fitBikeType, setFitBikeType] = useState<BikeType>('fixed');
  const [fitResult, setFitResult] = useState<ReturnType<typeof recommendFrameSize>>(null);

  const load = () => getListings(filters).then(setListings);
  const selectedLabels = (filters.labels ?? '').split(',').filter(Boolean);
  const toggleLabel = (label: string) => {
    const next = selectedLabels.includes(label)
      ? selectedLabels.filter((item) => item !== label)
      : [...selectedLabels, label];
    setFilters({ ...filters, labels: next.join(',') });
  };
  const calculateFit = () => setFitResult(recommendFrameSize(Number(fitHeight), fitBikeType));
  const showFitListings = () => {
    if (!fitResult) return;
    const nextFilters = {
      ...filters,
      category: '',
      bike_type: fitBikeType,
      frame_size: fitResult.frameSizes.join(',')
    };
    setFilters(nextFilters);
    getListings(nextFilters).then(setListings);
  };

  useEffect(() => {
    const category = params.get('category') ?? '';
    setFilters((current) => ({ ...current, category }));
    getListings({ ...filters, category }).then(setListings);
  }, [params]);

  return (
    <div className="space-y-8">
      <div className="flex flex-col justify-between gap-4 md:flex-row md:items-end">
        <div>
          <h1 className="text-5xl font-black uppercase">Каталог</h1>
          <p className="mt-2 text-white/60">Выбери раздел: объявления, услуги мастеров или события вело-сообщества.</p>
        </div>
        <button className="btn" onClick={load}><Search size={18} /> Искать</button>
      </div>

      <section className="grid gap-4 lg:grid-cols-3">
        <Link to="/bikes" className="group border border-white/10 bg-black p-6 shadow-street transition hover:-translate-y-1">
          <div className="text-sm font-black uppercase text-acid">Каталог</div>
          <h2 className="mt-3 text-4xl font-black uppercase">Marketplace</h2>
          <p className="mt-2 text-white/60">Велосипеды целиком, рамы, колёса, трансмиссия, тормоза, cockpit, посадка.</p>
          <div className="mt-5 grid grid-cols-2 gap-2 text-sm font-black uppercase">
            {['Bikes', 'Frames', 'Wheels', 'Parts'].map((item) => <span key={item} className="bg-steel px-3 py-2 group-hover:bg-acid group-hover:text-black">{item}</span>)}
          </div>
        </Link>
        <Link to="/services" className="group border border-white/10 bg-black p-6 shadow-danger transition hover:-translate-y-1">
          <div className="text-sm font-black uppercase text-danger">Каталог</div>
          <h2 className="mt-3 text-4xl font-black uppercase">Velo Services</h2>
          <p className="mt-2 text-white/60">Ремонт, покраска, сборка, настройка и доставка от мастеров.</p>
        </Link>
        <Link to="/community" className="group border border-white/10 bg-black p-6 transition hover:-translate-y-1 hover:shadow-street">
          <div className="text-sm font-black uppercase text-acid">Каталог</div>
          <h2 className="mt-3 text-4xl font-black uppercase">Velo Community</h2>
          <p className="mt-2 text-white/60">Покатушки, заезды, встречи и вело-комьюнити Кыргызстана.</p>
        </Link>
      </section>

      <section className="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
        {CATEGORIES.map((category) => (
          <Link key={category} to={categoryPath(category)} className="border border-white/15 bg-steel p-4 text-center font-black uppercase hover:bg-acid hover:text-black">
            {category}
          </Link>
        ))}
      </section>

      <section className="panel grid gap-5 overflow-hidden p-5 md:grid-cols-[1fr_auto] md:items-end">
        <div className="space-y-4">
          <div>
            <div className="inline-block bg-danger px-3 py-1 text-xs font-black uppercase text-white">Fit check</div>
            <h2 className="mt-3 text-4xl font-black uppercase">Подобрать ростовку</h2>
          </div>
          <div className="grid gap-3 md:grid-cols-3">
            <input className="field" inputMode="numeric" placeholder="Рост, см" value={fitHeight} onChange={(e) => setFitHeight(e.target.value)} />
            <select className="field" value={fitBikeType} onChange={(e) => setFitBikeType(e.target.value as BikeType)}>
              {BIKE_TYPES.map((type) => <option key={type.value} value={type.value}>{type.label}</option>)}
            </select>
            <button className="btn" type="button" onClick={calculateFit}>Подобрать</button>
          </div>
          {fitResult && (
            <div className="border border-acid bg-black p-4">
              <div className="text-sm font-black uppercase text-white/55">Результат</div>
              <div className="mt-1 text-3xl font-black uppercase text-acid">Тебе подойдёт {fitResult.label}</div>
              <p className="mt-2 text-sm text-white/55">Подбор примерный и зависит от геометрии рамы, длины ног, выноса и посадки.</p>
            </div>
          )}
        </div>
        {fitResult && <button className="btn-dark h-fit" type="button" onClick={showFitListings}>Показать подходящие объявления</button>}
      </section>

      <h2 className="text-3xl font-black uppercase">Все объявления</h2>
      <div className="panel grid gap-3 p-4 md:grid-cols-8">
        <input className="field md:col-span-2" placeholder="Поиск" value={filters.search ?? ''} onChange={(e) => setFilters({ ...filters, search: e.target.value })} />
        <input className="field" placeholder="Категория" value={filters.category ?? ''} onChange={(e) => setFilters({ ...filters, category: e.target.value })} />
        <input className="field" placeholder="Бренд" value={filters.brand ?? ''} onChange={(e) => setFilters({ ...filters, brand: e.target.value })} />
        <input className="field" placeholder="Город: Бишкек, Ош, Каракол..." value={filters.city ?? ''} onChange={(e) => setFilters({ ...filters, city: e.target.value })} />
        <input className="field" placeholder="Состояние" value={filters.condition ?? ''} onChange={(e) => setFilters({ ...filters, condition: e.target.value })} />
        <select className="field" value={filters.bike_type ?? ''} onChange={(e) => setFilters({ ...filters, bike_type: e.target.value })}>
          <option value="">Любой тип</option>
          {BIKE_TYPES.map((type) => <option key={type.value} value={type.value}>{type.label}</option>)}
        </select>
        <input className="field" placeholder="Ростовка S/M/L/54" value={filters.frame_size ?? ''} onChange={(e) => setFilters({ ...filters, frame_size: e.target.value })} />
        <input className="field" placeholder="Рост райдера, см" value={filters.rider_height ?? ''} onChange={(e) => setFilters({ ...filters, rider_height: e.target.value })} />
        <input className="field" placeholder="Цена от" value={filters.min_price ?? ''} onChange={(e) => setFilters({ ...filters, min_price: e.target.value })} />
        <input className="field" placeholder="Цена до" value={filters.max_price ?? ''} onChange={(e) => setFilters({ ...filters, max_price: e.target.value })} />
        <select className="field" value={filters.deal_type ?? ''} onChange={(e) => setFilters({ ...filters, deal_type: e.target.value })}>
          <option value="">Любая сделка</option>
          {DEAL_TYPES.map((dealType) => <option key={dealType}>{dealType}</option>)}
        </select>
        <select className="field" value={filters.sort ?? 'new'} onChange={(e) => setFilters({ ...filters, sort: e.target.value })}>
          <option value="new">Сначала новые</option>
          <option value="price_asc">Цена ↑</option>
          <option value="price_desc">Цена ↓</option>
          <option value="popular">Популярные</option>
          <option value="price_reduced_recently">Недавно снижена цена</option>
          <option value="biggest_reduction">Самое большое снижение</option>
          <option value="minimum_price">Минимальная цена</option>
          <option value="maximum_price">Максимальная цена</option>
        </select>
        <label className="flex items-center gap-3 border border-danger bg-black px-4 font-black uppercase text-danger"><input type="checkbox" checked={!!filters.price_reduced} onChange={(e) => setFilters({ ...filters, price_reduced: e.target.checked })}/> Цена снижена</label>
        <div className="grid gap-2 md:col-span-8 sm:grid-cols-2 lg:grid-cols-4">
          {LISTING_LABELS.map((label) => {
            const active = selectedLabels.includes(label);
            return (
              <button
                key={label}
                type="button"
                onClick={() => toggleLabel(label)}
                className={`border px-4 py-3 text-left font-black uppercase transition hover:-translate-y-1 ${
                  active ? labelTone(label) : 'border-white/15 bg-black text-white/55 hover:border-acid hover:text-acid'
                }`}
              >
                {label}
              </button>
            );
          })}
        </div>
      </div>
      <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
        {listings.map((listing) => <ListingCard key={listing.id} listing={listing} />)}
      </div>
    </div>
  );
}
