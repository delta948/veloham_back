import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { api } from '../api/client';
import { CATEGORIES } from '../constants/catalog';
import type { WantedRequest } from '../types';

export function WantedFormPage() {
  const navigate = useNavigate();
  const [form, setForm] = useState({
    title: '',
    category: 'Велосипеды целиком',
    min_budget: '',
    max_budget: '',
    city: '',
    frame_size: '',
    rider_height: '',
    description: '',
    status: 'active'
  });

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    const payload = {
      ...form,
      min_budget: Number(form.min_budget || 0),
      max_budget: Number(form.max_budget || 0),
      rider_height: Number(form.rider_height || 0)
    };
    const { data } = await api.post<WantedRequest>('/wanted', payload);
    navigate(`/wanted/${data.id}`);
  };

  return (
    <form onSubmit={submit} className="panel mx-auto max-w-3xl space-y-5 p-6">
      <h1 className="text-4xl font-black uppercase">Создать заявку</h1>
      <input className="field" placeholder="Что хочешь купить" value={form.title} onChange={(e) => setForm({ ...form, title: e.target.value })} />
      <textarea className="field min-h-32" placeholder="Описание" value={form.description} onChange={(e) => setForm({ ...form, description: e.target.value })} />
      <div className="grid gap-4 md:grid-cols-2">
        <select className="field" value={form.category} onChange={(e) => setForm({ ...form, category: e.target.value })}>
          {CATEGORIES.map((item) => <option key={item}>{item}</option>)}
        </select>
        <input className="field" placeholder="Город" value={form.city} onChange={(e) => setForm({ ...form, city: e.target.value })} />
        <input className="field" placeholder="Бюджет от" value={form.min_budget} onChange={(e) => setForm({ ...form, min_budget: e.target.value })} />
        <input className="field" placeholder="Бюджет до" value={form.max_budget} onChange={(e) => setForm({ ...form, max_budget: e.target.value })} />
        <input className="field" placeholder="Ростовка S/M/L или 54" value={form.frame_size} onChange={(e) => setForm({ ...form, frame_size: e.target.value })} />
        <input className="field" placeholder="Рост райдера, см" value={form.rider_height} onChange={(e) => setForm({ ...form, rider_height: e.target.value })} />
        <select className="field" value={form.status} onChange={(e) => setForm({ ...form, status: e.target.value })}>
          <option value="active">Активно</option>
          <option value="closed">Закрыто</option>
        </select>
      </div>
      <button className="btn w-full">Опубликовать заявку</button>
    </form>
  );
}
