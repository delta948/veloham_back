import { TrendingDown } from 'lucide-react';
import type { Listing } from '../types';
import { formatSom } from '../utils/format';

export function PriceDropBadge({ listing, detailed = false }: { listing: Listing; detailed?: boolean }) {
  if (!listing.previous_price || listing.price >= listing.previous_price) return null;
  const amount = listing.previous_price - listing.price;
  const base = listing.initial_price || listing.previous_price;
  if (listing.price >= base) return <span className="inline-flex items-center gap-1 border border-white/20 px-2 py-1 text-xs font-black uppercase text-white/60">Цена изменена</span>;
  const percent = Math.max(0, Math.round(((base - listing.price) / base) * 100));
  return <div className="space-y-1">
    <span className="inline-flex items-center gap-1 bg-danger px-2 py-1 text-xs font-black uppercase text-white"><TrendingDown size={14} /> Цена снижена</span>
    {detailed && <><div className="text-xl font-bold text-white/45 line-through">{formatSom(listing.previous_price)}</div><div className="text-sm font-black text-danger">Цена снижена на {formatSom(amount)} · −{percent}% от первоначальной</div></>}
  </div>;
}
