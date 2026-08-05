import { useEffect, useState } from 'react';
import { api } from '../api/client';
import type { AdminPriceHistory, Listing, Report, User } from '../types';

export function AdminPage() {
  const [users, setUsers] = useState<User[]>([]);
  const [listings, setListings] = useState<Listing[]>([]);
  const [reports, setReports] = useState<Report[]>([]);
  const [priceHistory, setPriceHistory] = useState<AdminPriceHistory[]>([]);

  const load = () => {
    api.get<User[]>('/admin/users').then(({ data }) => setUsers(data)).catch(() => setUsers([]));
    api.get<Listing[]>('/admin/listings').then(({ data }) => setListings(data)).catch(() => setListings([]));
    api.get<Report[]>('/admin/reports').then(({ data }) => setReports(data)).catch(() => setReports([]));
    api.get<AdminPriceHistory[]>('/admin/price-history').then(({ data }) => setPriceHistory(data)).catch(() => setPriceHistory([]));
  };

  useEffect(() => { load(); }, []);

  return (
    <div className="space-y-8">
      <h1 className="text-5xl font-black uppercase">Админ-панель</h1>
      <section className="grid gap-4 md:grid-cols-3">
        <div className="panel p-5"><div className="text-4xl font-black text-acid">{users.length}</div><div className="uppercase text-white/60">Пользователи</div></div>
        <div className="panel p-5"><div className="text-4xl font-black text-acid">{listings.length}</div><div className="uppercase text-white/60">Объявления</div></div>
        <div className="panel p-5"><div className="text-4xl font-black text-danger">{reports.length}</div><div className="uppercase text-white/60">Жалобы</div></div>
      </section>
      <section className="panel overflow-auto p-5">
        <h2 className="mb-4 text-3xl font-black uppercase">Пользователи</h2>
        {users.map((user) => (
          <div key={user.id} className="flex min-w-[720px] items-center justify-between border-t border-white/10 py-3">
            <span>{user.username} · {user.email} · {user.role}</span>
            <button className="btn-dark px-3 py-2" onClick={async () => { await api.patch(`/admin/users/${user.id}/block`, { is_blocked: !user.is_blocked }); load(); }}>{user.is_blocked ? 'Разблокировать' : 'Заблокировать'}</button>
          </div>
        ))}
      </section>
      <section className="panel overflow-auto p-5">
        <h2 className="mb-4 text-3xl font-black uppercase">Объявления</h2>
        {listings.map((listing) => (
          <div key={listing.id} className="flex min-w-[720px] items-center justify-between border-t border-white/10 py-3">
            <span>{listing.title} · {listing.category} · {listing.status}</span>
            <button className="btn-dark px-3 py-2 text-danger" onClick={async () => { await api.delete(`/admin/listings/${listing.id}`); load(); }}>Удалить</button>
          </div>
        ))}
      </section>
      <section className="panel p-5">
        <h2 className="mb-4 text-3xl font-black uppercase">История цен</h2>
        <div className="overflow-x-auto"><table className="min-w-[1000px] w-full text-left text-sm"><thead><tr className="text-acid"><th>Объявление / продавец</th><th>Цена</th><th>Дата / IP</th><th>Изменил</th><th>Всего</th><th>Риск</th></tr></thead><tbody>{priceHistory.map((row) => <tr key={row.id} className="border-t border-white/10"><td className="py-3">{row.listing?.title}<br/><span className="text-white/45">{row.listing?.user?.username}</span></td><td>{row.old_price.toLocaleString('ru-KG')} → {row.new_price.toLocaleString('ru-KG')}</td><td>{new Date(row.changed_at).toLocaleString('ru-KG')}<br/>{row.ip_address || '—'}</td><td>{row.changed_by_user?.username || row.changed_by}</td><td>{row.change_count}</td><td>{row.suspicious ? <span className="bg-danger px-2 py-1 font-black">Подозрительное изменение<br/>{row.suspicious_reason}</span> : '—'}</td></tr>)}</tbody></table></div>
      </section>
      <section className="panel p-5">
        <h2 className="mb-4 text-3xl font-black uppercase">Жалобы</h2>
        {reports.map((report) => (
          <div key={report.id} className="border-t border-white/10 py-3 text-white/75">
            <b className="text-danger">{report.reason}</b> · {report.text || 'без текста'} · {new Date(report.created_at).toLocaleString('ru-KG')}
          </div>
        ))}
      </section>
    </div>
  );
}
