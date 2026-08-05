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
    {open && <div className="absolute right-0 top-12 z-40 w-80 border border-white/15 bg-black p-2 shadow-street">
      {items.length === 0 ? <p className="p-3 text-sm text-white/50">Новых уведомлений нет.</p> : items.map((item) => <button key={item.id} className={`block w-full border-b border-white/10 p-3 text-left text-sm ${item.is_read ? 'text-white/45' : 'text-white'}`} onClick={async () => { await api.patch(`/notifications/${item.id}/read`); navigate(item.link); setOpen(false); }}>{item.message}</button>)}
    </div>}
  </div>;
}
