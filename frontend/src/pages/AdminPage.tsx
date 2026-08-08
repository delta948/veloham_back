import { useEffect, useState } from 'react';
import { api } from '../api/client';
import type { AdminPriceHistory, AdminStats, Listing, ListingPlacement, Report, User, UserBlockEvent } from '../types';

export function AdminPage() {
  const [users, setUsers] = useState<User[]>([]);
  const [listings, setListings] = useState<Listing[]>([]);
  const [reports, setReports] = useState<Report[]>([]);
  const [priceHistory, setPriceHistory] = useState<AdminPriceHistory[]>([]);
  const [blockReasons, setBlockReasons] = useState<Record<string, string>>({});
  const [userActionError, setUserActionError] = useState('');
  const [search, setSearch] = useState('');
  const [stats, setStats] = useState<AdminStats | null>(null);
  const [blockEvents, setBlockEvents] = useState<UserBlockEvent[]>([]);
  const [payments, setPayments] = useState<ListingPlacement[]>([]);

  const load = () => {
    api.get<User[]>('/admin/users', { params: search.trim() ? { q: search.trim() } : undefined }).then(({ data }) => setUsers(data)).catch(() => setUsers([]));
    api.get<Listing[]>('/admin/listings').then(({ data }) => setListings(data)).catch(() => setListings([]));
    api.get<Report[]>('/admin/reports').then(({ data }) => setReports(data)).catch(() => setReports([]));
    api.get<AdminPriceHistory[]>('/admin/price-history').then(({ data }) => setPriceHistory(data)).catch(() => setPriceHistory([]));
    api.get<AdminStats>('/admin/stats').then(({ data }) => setStats(data)).catch(() => setStats(null));
    api.get<UserBlockEvent[]>('/admin/block-events').then(({ data }) => setBlockEvents(data)).catch(() => setBlockEvents([]));
    api.get<ListingPlacement[]>('/admin/payments').then(({ data }) => setPayments(data)).catch(() => setPayments([]));
  };

  useEffect(() => { load(); }, []);

  const toggleBlock = async (user: User) => {
    const reason = (blockReasons[user.id] ?? '').trim();
    if (!user.is_blocked && !reason) {
      setUserActionError('Укажите причину блокировки.');
      return;
    }
    setUserActionError('');
    try {
      await api.patch(`/admin/users/${user.id}/block`, { is_blocked: !user.is_blocked, reason });
      setBlockReasons((current) => ({ ...current, [user.id]: '' }));
      load();
    } catch {
      setUserActionError('Не удалось изменить блокировку пользователя.');
    }
  };

  return (
    <div className="space-y-6 sm:space-y-8">
      <h1 className="break-words text-3xl font-black uppercase sm:text-5xl">Админ-панель</h1>
      <section className="grid gap-4 md:grid-cols-3">
        <div className="panel p-5"><div className="text-4xl font-black text-acid">{stats?.users ?? users.length}</div><div className="uppercase text-white/60">Пользователи · блок {stats?.blocked_users ?? 0}</div></div>
        <div className="panel p-5"><div className="text-4xl font-black text-acid">{stats?.listings ?? listings.length}</div><div className="uppercase text-white/60">Объявления</div></div>
        <div className="panel p-5"><div className="text-4xl font-black text-danger">{stats?.reports ?? reports.length}</div><div className="uppercase text-white/60">Жалобы · платежи {stats?.paid_payments ?? 0}</div></div>
      </section>
      <section className="panel p-4 sm:p-5">
        <h2 className="mb-4 text-2xl font-black uppercase sm:text-3xl">Пользователи</h2>
        <form className="mb-4 flex flex-col gap-2 sm:flex-row" onSubmit={(event) => { event.preventDefault(); load(); }}><input className="field" placeholder="Имя, email или телефон" value={search} onChange={(event) => setSearch(event.target.value)} /><button className="btn sm:w-auto">Найти</button></form>
        {userActionError && <div className="mb-3 bg-danger p-3 font-bold">{userActionError}</div>}
        {users.map((user) => (
          <div key={user.id} className="grid gap-3 border-t border-white/10 py-3 md:grid-cols-[minmax(0,1fr)_minmax(220px,1fr)_auto] md:items-center">
            <div className="min-w-0 break-words text-sm sm:text-base"><span>{user.username} · {user.phone || user.email || 'без контакта'} · {user.role}</span>{user.is_blocked && <p className="mt-1 text-sm text-danger">Заблокирован: {user.blocked_reason || 'причина не указана'}</p>}</div>
            {!user.is_blocked && <input className="field" maxLength={500} placeholder="Причина блокировки" value={blockReasons[user.id] ?? ''} onChange={(event) => setBlockReasons((current) => ({ ...current, [user.id]: event.target.value }))} />}
            <button className="btn-dark w-full px-3 py-2 md:w-auto" onClick={() => void toggleBlock(user)}>{user.is_blocked ? 'Разблокировать' : 'Заблокировать'}</button>
          </div>
        ))}
      </section>
      <section className="panel p-4 sm:p-5">
        <h2 className="mb-4 text-2xl font-black uppercase sm:text-3xl">Платежи</h2>
        {payments.length === 0 && <p className="text-white/55">Платежей пока нет.</p>}
        {payments.map((payment) => <div key={payment.id} className="flex flex-col gap-3 border-t border-white/10 py-3 sm:flex-row sm:items-center sm:justify-between"><div className="min-w-0 break-all text-sm"><b className="text-acid">{payment.amount} {payment.currency}</b> · {payment.status}<br/><span className="text-white/45">{payment.provider_payment_id || payment.id}</span></div>{payment.status !== 'paid' && payment.provider_payment_id && <button className="btn-dark w-full sm:w-auto" onClick={async () => { await api.post(`/admin/payments/${payment.id}/recheck`); load(); }}>Проверить оплату</button>}</div>)}
      </section>
      <section className="panel p-4 sm:p-5">
        <h2 className="mb-4 text-2xl font-black uppercase sm:text-3xl">Журнал блокировок</h2>
        {blockEvents.length === 0 && <p className="text-white/55">Действий пока нет.</p>}
        {blockEvents.map((event) => <div key={event.id} className="border-t border-white/10 py-3 text-sm"><b className={event.action === 'blocked' ? 'text-danger' : 'text-acid'}>{event.action === 'blocked' ? 'Заблокирован' : 'Разблокирован'}</b> · {event.user?.username || event.user_id}<div className="break-words text-white/55">{event.reason || 'без причины'} · администратор {event.admin?.username || event.admin_id} · {new Date(event.created_at).toLocaleString('ru-KG')}</div></div>)}
      </section>
      <section className="panel p-4 sm:p-5">
        <h2 className="mb-4 text-2xl font-black uppercase sm:text-3xl">Объявления</h2>
        {listings.map((listing) => (
          <div key={listing.id} className="flex flex-col gap-3 border-t border-white/10 py-3 sm:flex-row sm:items-center sm:justify-between">
            <span className="min-w-0 break-words text-sm sm:text-base">{listing.title} · {listing.category} · {listing.status}</span>
            <button className="btn-dark w-full px-3 py-2 text-danger sm:w-auto" onClick={async () => { await api.delete(`/admin/listings/${listing.id}`); load(); }}>Удалить</button>
          </div>
        ))}
      </section>
      <section className="panel p-4 sm:p-5">
        <h2 className="mb-4 text-2xl font-black uppercase sm:text-3xl">История цен</h2>
        <div className="grid gap-3 md:hidden">{priceHistory.map((row) => <article key={row.id} className="space-y-2 border border-white/10 bg-black p-3 text-sm"><b className="block break-words text-acid">{row.listing?.title || 'Объявление'}</b><div>{row.old_price.toLocaleString('ru-KG')} → {row.new_price.toLocaleString('ru-KG')} сом</div><div className="break-words text-white/50">{new Date(row.changed_at).toLocaleString('ru-KG')} · {row.ip_address || 'IP —'}</div><div>Изменил: {row.changed_by_user?.username || row.changed_by}</div>{row.suspicious && <div className="bg-danger p-2 font-black">Подозрительное изменение: {row.suspicious_reason}</div>}</article>)}</div>
        <div className="hidden overflow-x-auto md:block"><table className="w-full min-w-[1000px] text-left text-sm"><thead><tr className="text-acid"><th>Объявление / продавец</th><th>Цена</th><th>Дата / IP</th><th>Изменил</th><th>Всего</th><th>Риск</th></tr></thead><tbody>{priceHistory.map((row) => <tr key={row.id} className="border-t border-white/10"><td className="py-3">{row.listing?.title}<br/><span className="text-white/45">{row.listing?.user?.username}</span></td><td>{row.old_price.toLocaleString('ru-KG')} → {row.new_price.toLocaleString('ru-KG')}</td><td>{new Date(row.changed_at).toLocaleString('ru-KG')}<br/>{row.ip_address || '—'}</td><td>{row.changed_by_user?.username || row.changed_by}</td><td>{row.change_count}</td><td>{row.suspicious ? <span className="bg-danger px-2 py-1 font-black">Подозрительное изменение<br/>{row.suspicious_reason}</span> : '—'}</td></tr>)}</tbody></table></div>
      </section>
      <section className="panel p-4 sm:p-5">
        <h2 className="mb-4 text-2xl font-black uppercase sm:text-3xl">Жалобы</h2>
        {reports.map((report) => (
          <div key={report.id} className="border-t border-white/10 py-3 text-white/75">
            <b className="text-danger">{report.reason}</b> · {report.text || 'без текста'} · {new Date(report.created_at).toLocaleString('ru-KG')}
          </div>
        ))}
      </section>
    </div>
  );
}
