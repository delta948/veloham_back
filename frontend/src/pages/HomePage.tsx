import { motion, useReducedMotion, useScroll, useTransform } from 'framer-motion';
import { ArrowRight, Bike, Search, Sparkles } from 'lucide-react';
import { useEffect, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { api } from '../api/client';
import { getListings } from '../api/listings';
import { ActivityStatsWidget, BikeFinderWidget, BuyerRequestsWidget, CompareListingsWidget, DreamBuildWidget, NearbyListingsWidget, NotificationsWidget, PopularListingsWidget, PriceHistoryWidget, ProfileProgressWidget, RecentMessagesWidget, RecentlyViewedWidget, TopSellersWidget, UrgentListingsWidget } from '../components/home/LuxuryWidgets';
import { CATEGORIES, categoryPath } from '../constants/catalog';
import { useAuthStore } from '../store/auth';
import type { Listing, WantedRequest } from '../types';

export function HomePage() {
  const [listings,setListings]=useState<Listing[]>([]); const [requests,setRequests]=useState<WantedRequest[]>([]); const [loading,setLoading]=useState(true); const [error,setError]=useState(''); const [query,setQuery]=useState('');
  const user=useAuthStore(s=>s.user); const navigate=useNavigate(); const reduced=useReducedMotion(); const {scrollY}=useScroll(); const heroY=useTransform(scrollY,[0,650],[0,reduced?0:55]);
  useEffect(()=>{Promise.all([getListings(),api.get<WantedRequest[]>('/wanted').then(r=>r.data)]).then(([l,w])=>{setListings(l);setRequests(w)}).catch(()=>setError('Не удалось загрузить ленту VELOHAM')).finally(()=>setLoading(false))},[]);
  const reducedListing=listings.find(x=>x.previous_price&&x.price<x.previous_price)||listings[0]; const users=Array.from(new Map(listings.map(x=>[x.user?.id,x.user])).values()).filter(Boolean);
  const search=(e:React.FormEvent)=>{e.preventDefault();navigate(`/market?search=${encodeURIComponent(query)}`)};
  return <div className="space-y-7 pb-10 sm:space-y-10">
    <section className="hero-premium relative -mx-4 -mt-8 min-h-[690px] overflow-hidden border-b border-white/10 px-4 sm:min-h-[720px]">
      <motion.img style={{y:heroY}} src="/assets/veloham-hero-premium.jpg" alt="Премиальный велосипед VELOHAM" className="absolute inset-0 h-[110%] w-full object-cover object-center"/>
      <div className="absolute inset-0 bg-[linear-gradient(90deg,rgba(3,4,4,.98)_0%,rgba(3,4,4,.78)_38%,rgba(3,4,4,.2)_72%,rgba(3,4,4,.62)_100%)]"/><div className="hero-noise absolute inset-0 opacity-30"/><div className="hero-sweep absolute inset-0"/>
      <motion.div initial={reduced?false:{opacity:0,y:24}} animate={{opacity:1,y:0}} transition={{duration:.5}} className="relative z-10 mx-auto flex min-h-[690px] max-w-7xl items-center py-20 sm:min-h-[720px]">
        <div className="max-w-3xl"><div className="mb-5 inline-flex items-center gap-2 rounded-full border border-acid/30 bg-acid/[.06] px-3 py-2 text-xs font-black uppercase tracking-[.18em] text-acid"><Sparkles size={14}/> Веломаркет нового поколения</div><h1 className="text-[clamp(4.5rem,14vw,10rem)] font-black uppercase leading-[.72] tracking-[-.08em]">VELO<span className="text-acid">HAM</span></h1><p className="mt-8 text-2xl font-black uppercase tracking-tight text-white sm:text-4xl">Купи. Продай. Обменяй.</p><p className="mt-3 max-w-xl text-base leading-relaxed text-white/50 sm:text-lg">Велосипеды и компоненты от райдеров Кыргызстана. Быстро, честно, технологично.</p>
          <form onSubmit={search} className="mt-8 flex max-w-2xl rounded-2xl border border-white/15 bg-black/60 p-2 backdrop-blur-xl focus-within:border-acid/60"><Search className="ml-3 self-center text-white/30"/><input value={query} onChange={e=>setQuery(e.target.value)} className="min-w-0 flex-1 bg-transparent px-4 py-3 outline-none placeholder:text-white/30" placeholder="Найти велосипед, раму или колёса"/><button className="premium-action hidden sm:inline-flex">Найти</button></form>
          <div className="mt-4 flex flex-wrap gap-2">{CATEGORIES.slice(0,5).map(x=><Link className="hero-chip" to={categoryPath(x)} key={x}>{x}</Link>)}</div><div className="mt-8 flex flex-wrap gap-3"><Link className="premium-action" to="/create">Разместить объявление <ArrowRight size={17}/></Link><Link className="premium-secondary" to="/market"><Bike size={17}/> Найти велосипед</Link></div>
        </div>
      </motion.div>
    </section>

    {error&&<div className="rounded-2xl border border-danger/30 bg-danger/10 p-4 text-danger">{error}</div>}
    <div className="dashboard-grid">
      <PopularListingsWidget listings={listings} loading={loading} error={error}/>
      <PriceHistoryWidget listing={reducedListing}/>
      <UrgentListingsWidget listings={listings}/>
      <NearbyListingsWidget listings={listings} city={user?.city||'Бишкек'}/>
      <BikeFinderWidget listings={listings}/>
      <DreamBuildWidget/>
      <ActivityStatsWidget/>
      <TopSellersWidget users={users}/>
      <NotificationsWidget/>
      {user&&<RecentMessagesWidget/>}
      <BuyerRequestsWidget requests={requests}/>
      <CompareListingsWidget listings={listings}/>
      <RecentlyViewedWidget listings={listings}/>
      <ProfileProgressWidget user={user}/>
    </div>
    <footer className="border-t border-white/10 py-8 text-center text-xs font-black uppercase tracking-[.24em] text-white/25">VELOHAM · Кыргызстан · Ride further</footer>
  </div>;
}
