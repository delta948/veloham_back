import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { api } from '../api/client';
import type { Chat } from '../types';

export function ChatsPage() {
  const [chats, setChats] = useState<Chat[]>([]);
  useEffect(() => { api.get<Chat[]>('/chats').then(({ data }) => setChats(data)); }, []);

  return (
    <div className="space-y-6">
      <h1 className="text-5xl font-black uppercase">Чаты</h1>
      <div className="grid gap-4">
        {chats.map((chat) => (
          <Link key={chat.id} to={`/chats/${chat.id}`} className="panel flex items-center justify-between p-4 transition hover:border-acid">
            <div>
              <h2 className="text-xl font-black uppercase">{chat.listing?.title}</h2>
              <p className="text-white/60">Покупатель: {chat.buyer?.username} · Продавец: {chat.seller?.username}</p>
            </div>
            <span className="text-acid font-black">Открыть</span>
          </Link>
        ))}
      </div>
    </div>
  );
}
