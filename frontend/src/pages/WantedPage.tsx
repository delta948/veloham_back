import { Plus } from 'lucide-react';
import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { api } from '../api/client';
import type { WantedRequest } from '../types';
import { formatSom } from '../utils/format';

export function WantedPage() {
  const [items, setItems] = useState<WantedRequest[]>([]);

  useEffect(() => {
    api.get<WantedRequest[]>('/wanted').then(({ data }) => setItems(data));
  }, []);

  return (
    <div className="space-y-6 sm:space-y-8">
      <section className="flex flex-col justify-between gap-4 md:flex-row md:items-end">
        <div>
          <h1 className="text-3xl font-black uppercase sm:text-5xl">Хочу купить</h1>
          <p className="mt-2 text-white/60">Заявки покупателей: бюджет, город, ростовка и подходящие объявления.</p>
        </div>
        <Link className="btn w-full md:w-auto" to="/wanted/create"><Plus size={18} /> Создать заявку</Link>
      </section>

      <div className="grid gap-5 md:grid-cols-2 lg:grid-cols-3">
        {items.map((item) => (
          <Link key={item.id} to={`/wanted/${item.id}`} className="panel p-5 transition hover:-translate-y-1 hover:shadow-street">
            <div className="flex flex-wrap gap-2">
              <span className="bg-acid px-2 py-1 text-xs font-black uppercase text-black">{item.category}</span>
              <span className="bg-black px-2 py-1 text-xs font-black uppercase text-white/70">{item.status === 'closed' ? 'Закрыто' : 'Активно'}</span>
            </div>
            <h2 className="mt-4 text-2xl font-black uppercase">{item.title}</h2>
            <p className="mt-2 text-white/60">{item.city} · ростовка {item.frame_size || 'любая'} · рост {item.rider_height || 'не указан'} см</p>
            <div className="mt-4 text-2xl font-black text-acid">{formatSom(item.min_budget)} - {formatSom(item.max_budget)}</div>
          </Link>
        ))}
      </div>
    </div>
  );
}
