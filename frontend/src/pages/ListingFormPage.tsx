import { useEffect, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { getListing, saveListing } from '../api/listings';
import { CATEGORIES } from '../constants/catalog';
import { DEAL_TYPES, LISTING_LABELS, labelTone } from '../constants/listingMeta';
import { BIKE_TYPES, FRAME_SIZE_OPTIONS, type BikeType } from '../utils/sizeFit';

const conditions = ['новое', 'отличное', 'хорошее', 'требует ремонта'];

const empty = {
  title: '', description: '', price: '', city: '', brand: '', category: 'Велосипеды целиком', condition: 'хорошее',
  bike_type: 'fixed' as BikeType, frame_size: '', rider_height_min: '', rider_height_max: '', deal_type: 'продажа', labels: [] as string[], status: 'active',
  build_frame: '', build_size: '', build_fork: '', build_wheels: '', build_hubs: '', build_tires: '', build_handlebar: '', build_stem: '',
  build_saddle: '', build_cranks: '', build_bottom_bracket: '', build_chain: '', build_cog: '', build_brakes: '', build_weight: '',
  build_frame_condition: '', build_defects: '', build_documents: 'false',
  exchange_enabled: 'false', match_wants: '', match_categories: '', match_min_price: '', match_max_price: '', match_can_add_cash: 'false',
  match_max_cash_add: '', match_same_city_only: 'false'
};

export function ListingFormPage() {
  const { id } = useParams();
  const navigate = useNavigate();
  const [form, setForm] = useState(empty);
  const [files, setFiles] = useState<FileList | null>(null);

  useEffect(() => {
    if (id) getListing(id).then((listing) => setForm({
      title: listing.title,
      description: listing.description,
      price: String(listing.price),
      city: listing.city,
      brand: listing.brand ?? '',
      category: listing.category,
      bike_type: (listing.bike_type as BikeType) || 'fixed',
      condition: listing.condition,
      frame_size: listing.frame_size ?? '',
      rider_height_min: String(listing.rider_height_min ?? ''),
      rider_height_max: String(listing.rider_height_max ?? ''),
      deal_type: listing.deal_type ?? 'продажа',
      labels: listing.labels ?? [],
      status: listing.status,
      build_frame: listing.build_card?.frame ?? '',
      build_size: listing.build_card?.size ?? '',
      build_fork: listing.build_card?.fork ?? '',
      build_wheels: listing.build_card?.wheels ?? '',
      build_hubs: listing.build_card?.hubs ?? '',
      build_tires: listing.build_card?.tires ?? '',
      build_handlebar: listing.build_card?.handlebar ?? '',
      build_stem: listing.build_card?.stem ?? '',
      build_saddle: listing.build_card?.saddle ?? '',
      build_cranks: listing.build_card?.cranks ?? '',
      build_bottom_bracket: listing.build_card?.bottom_bracket ?? '',
      build_chain: listing.build_card?.chain ?? '',
      build_cog: listing.build_card?.cog ?? '',
      build_brakes: listing.build_card?.brakes ?? '',
      build_weight: listing.build_card?.weight ?? '',
      build_frame_condition: listing.build_card?.frame_condition ?? '',
      build_defects: listing.build_card?.defects ?? '',
      build_documents: String(Boolean(listing.build_card?.documents)),
      exchange_enabled: String(Boolean(listing.match_preference?.exchange_enabled)),
      match_wants: listing.match_preference?.wants ?? '',
      match_categories: listing.match_preference?.categories ?? '',
      match_min_price: String(listing.match_preference?.min_price ?? ''),
      match_max_price: String(listing.match_preference?.max_price ?? ''),
      match_can_add_cash: String(Boolean(listing.match_preference?.can_add_cash)),
      match_max_cash_add: String(listing.match_preference?.max_cash_add ?? ''),
      match_same_city_only: String(Boolean(listing.match_preference?.same_city_only))
    }));
  }, [id]);

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    const data = new FormData();
    Object.entries(form).forEach(([key, value]) => {
      if (key === 'labels') return;
      data.append(key, String(value));
    });
    form.labels.forEach((label) => data.append('labels', label));
    Array.from(files ?? []).forEach((file) => data.append('images', file));
    const listing = await saveListing(data, id);
    navigate(`/listing/${listing.id}`);
  };

  const toggleLabel = (label: string) => {
    setForm((current) => ({
      ...current,
      labels: current.labels.includes(label)
        ? current.labels.filter((item) => item !== label)
        : [...current.labels, label]
    }));
  };

  const frameSizeOptions = FRAME_SIZE_OPTIONS[form.bike_type];

  return (
    <form onSubmit={submit} className="panel mx-auto max-w-5xl space-y-8 p-6">
      <h1 className="text-4xl font-black uppercase">{id ? 'Редактировать' : 'Создать объявление'}</h1>
      <section className="space-y-4">
        <h2 className="text-2xl font-black uppercase text-acid">Объявление</h2>
        <input className="field" placeholder="Название" value={form.title} onChange={(e) => setForm({ ...form, title: e.target.value })} />
        <textarea className="field min-h-36" placeholder="Описание" value={form.description} onChange={(e) => setForm({ ...form, description: e.target.value })} />
        <div className="grid gap-4 md:grid-cols-3">
          <input className="field" placeholder="Цена в сомах" value={form.price} onChange={(e) => setForm({ ...form, price: e.target.value })} />
          <input className="field" placeholder="Город в Кыргызстане" value={form.city} onChange={(e) => setForm({ ...form, city: e.target.value })} />
          <input className="field" placeholder="Бренд" value={form.brand} onChange={(e) => setForm({ ...form, brand: e.target.value })} />
          <select className="field" value={form.category} onChange={(e) => setForm({ ...form, category: e.target.value })}>
            {CATEGORIES.map((v) => <option key={v}>{v}</option>)}
          </select>
          <select className="field" value={form.bike_type} onChange={(e) => setForm({ ...form, bike_type: e.target.value as BikeType, frame_size: '' })}>
            {BIKE_TYPES.map((type) => <option key={type.value} value={type.value}>{type.label}</option>)}
          </select>
          <select className="field" value={form.frame_size} onChange={(e) => setForm({ ...form, frame_size: e.target.value })} required>
            <option value="">Ростовка</option>
            {frameSizeOptions.map((size) => <option key={size} value={size}>{size}</option>)}
          </select>
          <input className="field" placeholder="Рост райдера от, см" value={form.rider_height_min} onChange={(e) => setForm({ ...form, rider_height_min: e.target.value })} />
          <input className="field" placeholder="Рост райдера до, см" value={form.rider_height_max} onChange={(e) => setForm({ ...form, rider_height_max: e.target.value })} />
          <select className="field" value={form.condition} onChange={(e) => setForm({ ...form, condition: e.target.value })}>
            {conditions.map((v) => <option key={v}>{v}</option>)}
          </select>
          <select className="field" value={form.deal_type} onChange={(e) => setForm({ ...form, deal_type: e.target.value })}>
            {DEAL_TYPES.map((v) => <option key={v}>{v}</option>)}
          </select>
          <select className="field" value={form.status} onChange={(e) => setForm({ ...form, status: e.target.value })}>
            <option value="active">активно</option>
            <option value="sold">продано</option>
            <option value="hidden">скрыто</option>
          </select>
          <input className="field md:col-span-3" type="file" multiple accept="image/*" onChange={(e) => setFiles(e.target.files)} />
        </div>
        <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
          {LISTING_LABELS.map((label) => {
            const active = form.labels.includes(label);
            return (
              <button
                key={label}
                type="button"
                onClick={() => toggleLabel(label)}
                className={`border px-4 py-3 text-left font-black uppercase transition hover:-translate-y-1 ${
                  active ? labelTone(label) : 'border-white/15 bg-black text-white/55 hover:border-acid hover:text-acid'
                }`}
              >
                {label}
              </button>
            );
          })}
        </div>
      </section>

      <section className="space-y-4">
        <h2 className="text-2xl font-black uppercase text-acid">Build Card</h2>
        <div className="grid gap-4 md:grid-cols-3">
          {[
            ['build_frame', 'Рама'], ['build_size', 'Ростовка'], ['build_fork', 'Вилка'], ['build_wheels', 'Колеса'], ['build_hubs', 'Втулки'], ['build_tires', 'Покрышки'],
            ['build_handlebar', 'Руль'], ['build_stem', 'Вынос'], ['build_saddle', 'Седло'], ['build_cranks', 'Шатуны'], ['build_bottom_bracket', 'Каретка'], ['build_chain', 'Цепь'],
            ['build_cog', 'Cog / звезда'], ['build_brakes', 'Тормоза'], ['build_weight', 'Вес'], ['build_frame_condition', 'Состояние рамы'], ['build_defects', 'Дефекты']
          ].map(([key, label]) => (
            <input key={key} className="field" placeholder={label} value={form[key as keyof typeof form]} onChange={(e) => setForm({ ...form, [key]: e.target.value })} />
          ))}
          <label className="flex items-center gap-3 border border-white/15 bg-black px-4 py-3 font-bold uppercase">
            <input type="checkbox" checked={form.build_documents === 'true'} onChange={(e) => setForm({ ...form, build_documents: String(e.target.checked) })} />
            Документы есть
          </label>
        </div>
      </section>

      <section className="space-y-4">
        <h2 className="text-2xl font-black uppercase text-acid">Bike Match</h2>
        <label className="flex items-center gap-3 border border-white/15 bg-black px-4 py-3 font-bold uppercase">
          <input type="checkbox" checked={form.exchange_enabled === 'true'} onChange={(e) => setForm({ ...form, exchange_enabled: String(e.target.checked) })} />
          Интересует обмен
        </label>
        <div className="grid gap-4 md:grid-cols-3">
          <input className="field md:col-span-2" placeholder="Что хочу взамен" value={form.match_wants} onChange={(e) => setForm({ ...form, match_wants: e.target.value })} />
          <input className="field" placeholder="Категории: Велосипеды целиком,Колёса" value={form.match_categories} onChange={(e) => setForm({ ...form, match_categories: e.target.value })} />
          <input className="field" placeholder="Мин. цена обмена" value={form.match_min_price} onChange={(e) => setForm({ ...form, match_min_price: e.target.value })} />
          <input className="field" placeholder="Макс. цена обмена" value={form.match_max_price} onChange={(e) => setForm({ ...form, match_max_price: e.target.value })} />
          <input className="field" placeholder="Макс. доплата" value={form.match_max_cash_add} onChange={(e) => setForm({ ...form, match_max_cash_add: e.target.value })} />
          <label className="flex items-center gap-3 border border-white/15 bg-black px-4 py-3 font-bold uppercase">
            <input type="checkbox" checked={form.match_can_add_cash === 'true'} onChange={(e) => setForm({ ...form, match_can_add_cash: String(e.target.checked) })} />
            Готов доплатить
          </label>
          <label className="flex items-center gap-3 border border-white/15 bg-black px-4 py-3 font-bold uppercase">
            <input type="checkbox" checked={form.match_same_city_only === 'true'} onChange={(e) => setForm({ ...form, match_same_city_only: String(e.target.checked) })} />
            Только свой город
          </label>
        </div>
      </section>
      <button className="btn w-full">{id ? 'Сохранить' : 'Опубликовать'}</button>
    </form>
  );
}
