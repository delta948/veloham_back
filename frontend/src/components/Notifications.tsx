import { Bell } from 'lucide-react';
import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { api } from '../api/client';
import type { Notification } from '../types';

export function Notifications() {
  const [items, setItems] = useState<Notification[]>([]); const [open, setOpen] = useState(false); const navigate = useNavigate();
  useEffect(() => { api.get<Notification[]>('/notifications').then(({data}) => setItems(data)).catch(() => setItems([])); }, []);
  const unread = items.filter((x) => !x.is_read).length;
  return <div className="relative"><button title="Уведомления" className="btn-dark relative px-3 py-2" onClick={() => setOpen(!open)}><Bell size={18}/>{unread > 0 && <b className="absolute -right-2 -top-2 bg-danger px-1 text-xs">{unread}</b>}</button>
    {open && <div className="fixed inset-x-3 top-[4.5rem] z-40 max-h-[70dvh] overflow-y-auto border border-white/15 bg-black p-2 shadow-street sm:absolute sm:inset-x-auto sm:right-0 sm:top-12 sm:w-80">
      {unread > 0 && <button className="w-full border-b border-white/10 p-2 text-right text-xs text-acid" onClick={async () => { await api.patch('/notifications/read-all'); setItems((current) => current.map((item) => ({ ...item, is_read: true }))); }}>Прочитать все</button>}
      {items.length === 0 ? <p className="p-3 text-sm text-white/50">Новых уведомлений нет.</p> : items.map((item) => <button key={item.id} className={`block w-full border-b border-white/10 p-3 text-left text-sm ${item.is_read ? 'text-white/45' : 'text-white'}`} onClick={async () => { await api.patch(`/notifications/${item.id}/read`); setItems((current) => current.map((row) => row.id === item.id ? { ...row, is_read: true } : row)); navigate(item.link); setOpen(false); }}>{item.message}</button>)}
    </div>}
  </div>;
}
