import { Navigate, Route, Routes } from 'react-router-dom';
import { lazy, Suspense, type ReactNode } from 'react';
import { Layout } from '../components/Layout';
import { MarketPage } from '../pages/MarketPage';
import { ListingPage } from '../pages/ListingPage';
import { ListingFormPage } from '../pages/ListingFormPage';
import { LoginPage } from '../pages/LoginPage';
import { RegisterPage } from '../pages/RegisterPage';
import { ProfilePage } from '../pages/ProfilePage';
import { FavoritesPage } from '../pages/FavoritesPage';
import { ChatsPage } from '../pages/ChatsPage';
import { ChatPage } from '../pages/ChatPage';
import { CategoryListingsPage } from '../pages/CategoryListingsPage';
import { SellerProfilePage } from '../pages/SellerProfilePage';
import { AdminPage } from '../pages/AdminPage';
import { WantedPage } from '../pages/WantedPage';
import { WantedFormPage } from '../pages/WantedFormPage';
import { WantedDetailPage } from '../pages/WantedDetailPage';
import { useAuthStore } from '../store/auth';

const HomePage = lazy(() => import('../pages/HomePage').then((module) => ({ default: module.HomePage })));

function Protected({ children }: { children: ReactNode }) {
  return useAuthStore((s) => s.token) ? children : <Navigate to="/login" replace />;
}

export function AppRouter() {
  return (
    <Layout>
      <Routes>
        <Route path="/" element={<Suspense fallback={<div className="grid min-h-[60vh] place-items-center"><div className="h-10 w-10 animate-spin rounded-full border-2 border-white/10 border-t-acid"/></div>}><HomePage /></Suspense>} />
        <Route path="/market" element={<MarketPage />} />
        <Route path="/wanted" element={<WantedPage />} />
        <Route path="/wanted/create" element={<Protected><WantedFormPage /></Protected>} />
        <Route path="/wanted/:id" element={<WantedDetailPage />} />
        <Route path="/bikes" element={<CategoryListingsPage category="Велосипеды целиком" title="Велосипеды целиком" subtitle="Полные велосипеды VELOHAM: fixed gear, road, MTB, BMX и городские сборки." />} />
        <Route path="/fixed" element={<CategoryListingsPage category="Fixed Gear" title="Фиксеры" subtitle="Fixed Gear объявления по Кыргызстану: рамы, комплиты, вилсеты, рули, звезды и street-сборки." />} />
        <Route path="/road" element={<CategoryListingsPage category="Road Bike" title="Road" subtitle="Шоссейники, endurance, aero, карбон, колеса, групсеты и все для быстрых асфальтовых выездов." />} />
        <Route path="/mtb" element={<CategoryListingsPage category="MTB" title="MTB" subtitle="Горные велосипеды, трейл, эндуро, хардтейлы, вилки, колеса и комплектующие для гор." />} />
        <Route path="/bmx" element={<CategoryListingsPage category="BMX" title="BMX" subtitle="BMX комплиты, рамы, вилки, рули, пеги, драйверы и железо для street, park и dirt." />} />
        <Route path="/framesets" element={<CategoryListingsPage category="Рамы / фреймсеты" title="Рамы / фреймсеты" subtitle="Рамы, фреймсеты, вилки и база для новой сборки." />} />
        <Route path="/wheels" element={<CategoryListingsPage category="Колёса" title="Колёса" subtitle="Вилсеты, обода, втулки, спицы, покрышки и колеса под разные стили катания." />} />
        <Route path="/drivetrain" element={<CategoryListingsPage category="Трансмиссия" title="Трансмиссия" subtitle="Шатуны, цепи, кассеты, звезды, переключатели, каретки и привод." />} />
        <Route path="/cockpit" element={<CategoryListingsPage category="Руль и управление" title="Руль и управление" subtitle="Рули, выносы, грипсы, обмотки, рулевые и контроль байка." />} />
        <Route path="/brakes" element={<CategoryListingsPage category="Тормоза" title="Тормоза" subtitle="Калиперы, ротора, ручки, гидролинии, колодки и тормозные апгрейды." />} />
        <Route path="/fit" element={<CategoryListingsPage category="Седло и посадка" title="Седло и посадка" subtitle="Седла, подседелы, зажимы и детали посадки." />} />
        <Route path="/accessories" element={<CategoryListingsPage category="Аксессуары" title="Аксессуары" subtitle="Фонари, замки, сумки, фляги, инструменты и городские мелочи." />} />
        <Route path="/gear" element={<CategoryListingsPage category="Экипировка" title="Экипировка" subtitle="Шлемы, перчатки, очки, одежда и защита для города и гонок." />} />
        <Route path="/services" element={<CategoryListingsPage category="Вело-услуги" title="Velo Services" subtitle="Ремонт, покраска, сборка, настройка, доставка и услуги вело-мастеров." />} />
        <Route path="/community" element={<CategoryListingsPage category="Вело-события" title="Velo Community" subtitle="Покатушки, заезды, встречи и события вело-сообщества." />} />
        <Route path="/listing/:id" element={<ListingPage />} />
        <Route path="/create" element={<Protected><ListingFormPage /></Protected>} />
        <Route path="/edit/:id" element={<Protected><ListingFormPage /></Protected>} />
        <Route path="/login" element={<LoginPage />} />
        <Route path="/register" element={<RegisterPage />} />
        <Route path="/profile" element={<Protected><ProfilePage /></Protected>} />
        <Route path="/profile/:id" element={<SellerProfilePage />} />
        <Route path="/favorites" element={<Protected><FavoritesPage /></Protected>} />
        <Route path="/chats" element={<Protected><ChatsPage /></Protected>} />
        <Route path="/chats/:id" element={<Protected><ChatPage /></Protected>} />
        <Route path="/admin" element={<Protected><AdminPage /></Protected>} />
      </Routes>
    </Layout>
  );
}
