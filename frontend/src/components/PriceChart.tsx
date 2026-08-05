import type { PriceHistory } from '../types';

export function PriceChart({ data }: { data: PriceHistory }) {
  const points = [{ price: data.initial_price }, ...data.history.map((x) => ({ price: x.new_price }))];
  if (points.length < 2) return null;
  const min = Math.min(...points.map((x) => x.price)); const max = Math.max(...points.map((x) => x.price));
  const coords = points.map((p, i) => `${(i / (points.length - 1)) * 100},${90 - ((p.price - min) / Math.max(1, max - min)) * 75}`).join(' ');
  return <div className="overflow-hidden border border-white/10 bg-black p-3" aria-label="График изменения цены">
    <svg viewBox="0 0 100 100" preserveAspectRatio="none" className="h-40 w-full" role="img">
      <polyline points={coords} fill="none" stroke="currentColor" strokeWidth="2" vectorEffect="non-scaling-stroke" className="text-acid" />
      {coords.split(' ').map((xy, i) => { const [cx, cy] = xy.split(','); return <circle key={i} cx={cx} cy={cy} r="2" vectorEffect="non-scaling-stroke" className="fill-danger" />; })}
    </svg>
    <div className="flex justify-between text-xs text-white/45"><span>Публикация</span><span>Сейчас</span></div>
  </div>;
}
