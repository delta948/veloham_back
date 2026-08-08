import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { api } from '../api/client';
import type { Chat } from '../types';

export function ChatsPage() {
  const [chats, setChats] = useState<Chat[]>([]);
  useEffect(() => { api.get<Chat[]>('/chats').then(({ data }) => setChats(data)); }, []);

  return (
    <div className="space-y-6">
      <h1 className="text-3xl font-black uppercase sm:text-5xl">Чаты</h1>
      <div className="grid gap-4">
        {chats.map((chat) => (
          <Link key={chat.id} to={`/chats/${chat.id}`} className="panel flex flex-col items-start gap-3 p-4 transition hover:border-acid sm:flex-row sm:items-center sm:justify-between">
            <div className="min-w-0">
              <h2 className="text-xl font-black uppercase">{chat.listing?.title}</h2>
              <p className="break-words text-sm text-white/60 sm:text-base">Покупатель: {chat.buyer?.username} · Продавец: {chat.seller?.username}</p>
            </div>
            <span className="text-acid font-black">Открыть</span>
          </Link>
        ))}
      </div>
    </div>
  );
}
