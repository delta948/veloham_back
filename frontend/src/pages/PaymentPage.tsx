import { useEffect, useState } from 'react';
import { Link, useSearchParams } from 'react-router-dom';
import { api } from '../api/client';

type Placement = { status: 'pending' | 'paid' | 'failed'; listing_id?: string; checkout_url?: string };

export function PaymentPage() {
  const [params] = useSearchParams();
  const order = params.get('order');
  const [payment, setPayment] = useState<Placement | null>(null);
  const [error, setError] = useState('');
  const retry = async () => { if (!order) return; const { data } = await api.post<Placement>(`/payments/${order}/checkout`); if (data.checkout_url) window.location.assign(data.checkout_url); };

  useEffect(() => {
    if (!order) { setError('Номер платежа отсутствует.'); return; }
    let stopped = false;
    let timer: number | undefined;
    const check = async () => {
      try {
        const { data } = await api.get<Placement>(`/payments/${order}`);
        if (stopped) return;
        setPayment(data);
        if (data.status === 'pending') timer = window.setTimeout(check, 3000);
      } catch { if (!stopped) setError('Не удалось проверить оплату.'); }
    };
    void check();
    return () => { stopped = true; if (timer) window.clearTimeout(timer); };
  }, [order]);

  return <section className="panel mx-auto max-w-xl space-y-5 p-5 text-center sm:p-8">
    <h1 className="text-3xl font-black uppercase sm:text-5xl">Оплата размещения</h1>
    {error && <p className="text-danger">{error}</p>}
    {!error && (!payment || payment.status === 'pending') && <><div className="mx-auto h-12 w-12 animate-spin rounded-full border-2 border-white/15 border-t-acid"/><p className="text-white/65">Проверяем поступление 30 сом. Страница обновится автоматически.</p>{payment?.checkout_url && <a className="btn-dark w-full" href={payment.checkout_url}>Вернуться к оплате</a>}</>}
    {payment?.status === 'paid' && <><p className="text-acid">Оплата подтверждена. Объявление опубликовано.</p>{payment.listing_id && <Link className="btn w-full" to={`/listing/${payment.listing_id}`}>Открыть объявление</Link>}</>}
    {payment?.status === 'failed' && <><p className="text-danger">Платёж не прошёл.</p><button className="btn-dark w-full" onClick={retry}>Попробовать снова</button></>}
  </section>;
}
