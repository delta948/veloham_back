import { Search } from 'lucide-react';
import { useEffect, useState } from 'react';
import { getListings, type ListingFilters } from '../api/listings';
import { ListingCard } from '../components/ListingCard';
import { DEAL_TYPES, LISTING_LABELS, labelTone } from '../constants/listingMeta';
import type { Listing } from '../types';
import { BIKE_TYPES } from '../utils/sizeFit';

type CategoryListingsPageProps = {
  category: string;
  categories?: string[];
  title: string;
  subtitle: string;
};

export function CategoryListingsPage({ category, categories, title, subtitle }: CategoryListingsPageProps) {
  const [filters, setFilters] = useState<ListingFilters>({ category });
  const [listings, setListings] = useState<Listing[]>([]);
  const isCommunity = category === 'Вело-события';
  const selectedLabels = (filters.labels ?? '').split(',').filter(Boolean);

  const load = () => getListings(categories?.length ? { ...filters, category: undefined, categories: categories.join(',') } : { ...filters, category }).then(setListings);
  const toggleLabel = (label: string) => {
    const next = selectedLabels.includes(label)
      ? selectedLabels.filter((item) => item !== label)
      : [...selectedLabels, label];
    setFilters({ ...filters, labels: next.join(',') });
  };

  useEffect(() => {
    void load();
  }, [category, categories]);

  return (
    <div className="space-y-6 sm:space-y-8">
      <section className="border border-white/10 bg-black p-4 shadow-street sm:p-6 md:p-8">
        <div className="flex flex-col justify-between gap-5 md:flex-row md:items-end">
          <div>
            <div className="mb-3 inline-block bg-acid px-3 py-1 text-sm font-black uppercase text-black">{category}</div>
            <h1 className="break-words text-3xl font-black uppercase sm:text-5xl md:text-7xl">{title}</h1>
            <p className="mt-3 max-w-2xl text-white/65">{subtitle}</p>
          </div>
          {!isCommunity && <button className="btn w-full md:w-auto" onClick={load}><Search size={18} /> Искать</button>}
        </div>
      </section>

      {!isCommunity && (
        <div className="panel grid gap-3 p-3 sm:p-4 md:grid-cols-6">
          <input className="field md:col-span-2" placeholder={`Поиск внутри ${title}`} value={filters.search ?? ''} onChange={(e) => setFilters({ ...filters, search: e.target.value })} />
          <input className="field" placeholder="Город: Бишкек, Ош..." value={filters.city ?? ''} onChange={(e) => setFilters({ ...filters, city: e.target.value })} />
          <input className="field" placeholder="Цена до" value={filters.max_price ?? ''} onChange={(e) => setFilters({ ...filters, max_price: e.target.value })} />
          <select className="field" value={filters.bike_type ?? ''} onChange={(e) => setFilters({ ...filters, bike_type: e.target.value })}>
            <option value="">Любой тип</option>
            {BIKE_TYPES.map((type) => <option key={type.value} value={type.value}>{type.label}</option>)}
          </select>
          <input className="field" placeholder="Ростовка" value={filters.frame_size ?? ''} onChange={(e) => setFilters({ ...filters, frame_size: e.target.value })} />
          <select className="field" value={filters.deal_type ?? ''} onChange={(e) => setFilters({ ...filters, deal_type: e.target.value })}>
            <option value="">Любая сделка</option>
            {DEAL_TYPES.map((dealType) => <option key={dealType}>{dealType}</option>)}
          </select>
          <select className="field" value={filters.sort ?? 'new'} onChange={(e) => setFilters({ ...filters, sort: e.target.value })}>
            <option value="new">Новые</option>
            <option value="price_asc">Цена ↑</option>
            <option value="price_desc">Цена ↓</option>
            <option value="popular">Популярные</option>
          </select>
          <div className="grid gap-2 md:col-span-6 sm:grid-cols-2 lg:grid-cols-4">
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
      )}

      {listings.length > 0 ? (
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
          {listings.map((listing) => <ListingCard key={listing.id} listing={listing} />)}
        </div>
      ) : (
        <div className="panel p-8 text-center">
          <h2 className="text-3xl font-black uppercase">Пока пусто</h2>
          <p className="mt-2 text-white/60">В этой категории еще нет объявлений. Можно подать первое.</p>
        </div>
      )}
    </div>
  );
}
