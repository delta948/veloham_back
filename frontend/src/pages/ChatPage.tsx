import { Send } from 'lucide-react';
import { useEffect, useRef, useState } from 'react';
import { useParams } from 'react-router-dom';
import { api, WS_BASE } from '../api/client';
import { useAuthStore } from '../store/auth';
import type { Message } from '../types';

const quickMessages = ['Торг есть?', 'Обмен интересен?', 'Где можно посмотреть?', 'Актуально?', 'Могу забрать сегодня'];

export function ChatPage() {
  const { id } = useParams();
  const user = useAuthStore((s) => s.user);
  const token = useAuthStore((s) => s.token);
  const [messages, setMessages] = useState<Message[]>([]);
  const [text, setText] = useState('');
  const socket = useRef<WebSocket | null>(null);

  useEffect(() => {
    if (!id || !user?.id || !token) return;
    api.get<Message[]>(`/chats/${id}/messages`).then(({ data }) => setMessages(data));
    socket.current = new WebSocket(`${WS_BASE}/chats/${id}`, ['access_token', token]);
    socket.current.onmessage = (event) => setMessages((prev) => [...prev, JSON.parse(event.data)]);
    return () => socket.current?.close();
  }, [id, user?.id, token]);

  const send = (e: React.FormEvent) => {
    e.preventDefault();
    if (!text.trim()) return;
    socket.current?.send(JSON.stringify({ text }));
    setText('');
  };
  const sendQuick = (value: string) => {
    socket.current?.send(JSON.stringify({ text: value }));
  };

  return (
    <div className="panel mx-auto flex h-[72vh] max-w-4xl flex-col">
      <div className="border-b border-white/10 p-4">
        <h1 className="text-3xl font-black uppercase">Чат</h1>
      </div>
      <div className="flex-1 space-y-3 overflow-auto p-4">
        {messages.map((message) => {
          const own = message.sender_id === user?.id;
          return (
            <div key={message.id} className={`max-w-[80%] border p-3 ${own ? 'ml-auto border-acid bg-acid text-black' : 'border-white/15 bg-black text-white'}`}>
              <div className="text-xs font-black uppercase opacity-70">{message.sender?.username ?? 'user'}</div>
              <p>{message.text}</p>
              <div className="mt-1 text-[11px] opacity-60">{new Date(message.created_at).toLocaleTimeString('ru-KG', { hour: '2-digit', minute: '2-digit' })} · {message.is_read ? 'прочитано' : 'не прочитано'}</div>
            </div>
          );
        })}
      </div>
      <div className="flex gap-2 overflow-x-auto border-t border-white/10 p-3">
        {quickMessages.map((item) => <button key={item} className="min-w-fit border border-white/15 bg-black px-3 py-2 text-xs font-black uppercase text-white/75 hover:border-acid hover:text-acid" onClick={() => sendQuick(item)}>{item}</button>)}
      </div>
      <form onSubmit={send} className="flex gap-3 border-t border-white/10 p-4">
        <input className="field" placeholder="Сообщение" value={text} onChange={(e) => setText(e.target.value)} />
        <button className="btn" title="Отправить"><Send size={18} /></button>
      </form>
    </div>
  );
}
