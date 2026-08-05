import { Bike, CalendarDays, Heart, MessageSquare, Plus, Search, Shield, User, Wrench } from 'lucide-react';
import { Link, NavLink, useNavigate } from 'react-router-dom';
import { useEffect } from 'react';
import { useAuthStore } from '../store/auth';
import { Notifications } from './Notifications';

const nav = [
  { to: '/market', label: 'Маркет', icon: Bike },
  { to: '/wanted', label: 'Купить', icon: Search },
  { to: '/services', label: 'Услуги', icon: Wrench },
  { to: '/community', label: 'События', icon: CalendarDays },
  { to: '/create', label: 'Продать', icon: Plus },
  { to: '/favorites', label: 'Избранное', icon: Heart },
  { to: '/chats', label: 'Чаты', icon: MessageSquare },
  { to: '/profile', label: 'Профиль', icon: User }
];

export function Layout({ children }: { children: React.ReactNode }) {
  const { token, user, loadMe, logout } = useAuthStore();
  const navigate = useNavigate();
  const visibleNav = user?.role === 'admin' ? [...nav, { to: '/admin', label: 'Admin', icon: Shield }] : nav;

  useEffect(() => { void loadMe(); }, [loadMe, token]);

  return (
    <div className="min-h-screen street-grid">
      <header className="sticky top-0 z-20 border-b border-white/10 bg-black/85 backdrop-blur">
        <div className="mx-auto flex max-w-7xl items-center justify-between px-4 py-4">
          <Link to="/" className="text-3xl font-black tracking-tight text-white">
            VELO<span className="text-acid">HAM</span>
          </Link>
          <nav className="hidden items-center gap-2 md:flex">
            {visibleNav.map(({ to, label, icon: Icon }) => (
              <NavLink key={to} to={to} className={({ isActive }) => `flex items-center gap-2 px-3 py-2 text-sm font-bold uppercase ${isActive ? 'bg-acid text-black' : 'text-white/70 hover:text-acid'}`}>
                <Icon size={16} /> {label}
              </NavLink>
            ))}
          </nav>
          <div className="flex items-center gap-3">
            {user ? (
              <><Notifications/><button className="btn-dark px-3 py-2 text-xs" onClick={() => { logout(); navigate('/'); }}>Выйти</button></>
            ) : (
              <Link className="btn-dark px-3 py-2 text-xs" to="/login">Войти</Link>
            )}
          </div>
        </div>
        <nav className="flex overflow-x-auto border-t border-white/10 md:hidden">
          {visibleNav.map(({ to, label, icon: Icon }) => (
            <NavLink key={to} to={to} className={({ isActive }) => `flex min-w-[76px] flex-col items-center gap-1 px-1 py-2 text-[10px] font-bold uppercase ${isActive ? 'bg-acid text-black' : 'bg-black text-white/70'}`}>
              <Icon size={17} /> {label}
            </NavLink>
          ))}
        </nav>
      </header>
      <main className="mx-auto max-w-7xl px-4 py-8">{children}</main>
    </div>
  );
}
