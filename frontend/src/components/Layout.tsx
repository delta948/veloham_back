import { Bike, CalendarDays, Heart, LogOut, MessageSquare, Plus, Search, Shield, User, Wrench } from 'lucide-react';
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
  const mobileNav = nav.filter(({ to }) => ['/market', '/wanted', '/create', '/chats', '/profile'].includes(to));

  useEffect(() => { void loadMe(); }, [loadMe, token]);

  return (
    <div className="min-h-screen street-grid">
      <header className="mobile-safe-top sticky top-0 z-30 border-b border-white/10 bg-black/90 backdrop-blur">
        <div className="mx-auto flex max-w-7xl items-center justify-between px-3 py-3 sm:px-4 sm:py-4">
          <Link to="/" className="text-2xl font-black tracking-tight text-white sm:text-3xl">
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
              <>{user.role === 'admin' && <Link aria-label="Админ-панель" title="Админ-панель" className="btn-dark px-3 py-2 md:hidden" to="/admin"><Shield size={18}/></Link>}<Notifications/><button aria-label="Выйти" title="Выйти" className="btn-dark px-3 py-2 text-xs" onClick={() => { logout(); navigate('/'); }}><LogOut size={18} className="md:hidden"/><span className="hidden md:inline">Выйти</span></button></>
            ) : (
              <Link className="btn-dark px-3 py-2 text-xs" to="/login">Войти</Link>
            )}
          </div>
        </div>
        <nav aria-label="Мобильная навигация" className="mobile-safe-bottom fixed inset-x-0 bottom-0 z-50 grid grid-cols-5 border-t border-white/10 bg-black/95 backdrop-blur-xl md:hidden">
          {mobileNav.map(({ to, label, icon: Icon }) => (
            <NavLink key={to} to={to} className={({ isActive }) => `flex min-h-16 min-w-0 flex-col items-center justify-center gap-1 px-1 py-2 text-center text-[8px] font-bold uppercase min-[360px]:text-[9px] ${isActive ? 'bg-acid text-black' : 'text-white/70'}`}>
              <Icon size={17} /> {label}
            </NavLink>
          ))}
        </nav>
      </header>
      <main className="mx-auto max-w-7xl px-3 py-5 pb-24 sm:px-4 sm:py-8 md:pb-8">{children}</main>
    </div>
  );
}
