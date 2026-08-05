import { ChevronDown, TrendingDown, TrendingUp } from 'lucide-react';
import { useEffect, useState } from 'react';
import { getPriceHistory } from '../api/listings';
import type { PriceHistory as PriceHistoryData } from '../types';
import { formatSom } from '../utils/format';
import { PriceChart } from './PriceChart';

export function PriceHistory({ listingId, createdAt }: { listingId: string; createdAt: string }) {
  const [data, setData] = useState<PriceHistoryData | null>(null); const [open, setOpen] = useState(false); const [error, setError] = useState('');
  useEffect(() => { setError(''); getPriceHistory(listingId).then(setData).catch(() => setError('Не удалось загрузить историю цены.')); }, [listingId]);
  if (error) return <section className="panel border-danger p-5 text-danger">{error}<button className="btn-dark ml-3" onClick={() => location.reload()}>Повторить</button></section>;
  if (!data) return <section className="panel animate-pulse space-y-3 p-5"><div className="h-8 w-52 bg-white/10"/><div className="h-24 bg-white/5"/></section>;
  const visible = open ? data.history : data.history.slice(-1);
  return <section className="panel p-5">
    <h2 className="text-3xl font-black uppercase">История цены</h2>
    <div className="mt-5 border-l-2 border-acid pl-5">
      <div className="mb-5"><time className="text-sm text-white/45">{new Date(createdAt).toLocaleDateString('ru-KG')}</time><div className="font-black">{formatSom(data.initial_price)} — первоначальная цена</div></div>
      {data.history.length === 0 ? <p className="text-white/55">Цена объявления пока не изменялась.</p> : visible.map((item) => {
        const down = item.change_amount < 0; return <div key={item.changed_at} className="relative mb-5 before:absolute before:-left-[27px] before:top-2 before:h-3 before:w-3 before:bg-danger">
          <time className="text-sm text-white/45">{new Date(item.changed_at).toLocaleString('ru-KG')}</time>
          <div className="flex items-center gap-2 text-lg font-black">{formatSom(item.old_price)} → {formatSom(item.new_price)} {down ? <TrendingDown className="text-danger"/> : <TrendingUp className="text-white/45"/>}</div>
          <div className={down ? 'text-danger' : 'text-white/55'}>Цена {down ? 'снижена' : 'повышена'} на {formatSom(Math.abs(item.change_amount))}, или на {Math.abs(item.change_percent).toLocaleString('ru-KG')}%</div>
        </div>;
      })}
    </div>
    {data.history.length > 1 && <button className="btn-dark mb-5" onClick={() => setOpen(!open)}><ChevronDown className={open ? 'rotate-180' : ''}/> {open ? 'Скрыть полную историю' : 'Показать историю цены'}</button>}
    <PriceChart data={data}/><div className="mt-4 text-2xl font-black text-acid">Текущая цена: {formatSom(data.current_price)}</div>
  </section>;
}
