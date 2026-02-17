import React, { useState, useEffect, useCallback, Fragment } from 'react';
import { NavLink, Link, Outlet, useLocation } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import { FiGrid, FiUsers, FiTarget, FiMessageSquare, FiAward, FiBarChart2, FiLogOut, FiMenu, FiX, FiUser } from 'react-icons/fi';
import { Transition } from '@headlessui/react';
import api from '../services/api';
import { Match, APIResponse } from '../types';


const navItems = [
  { to: '/dashboard', label: 'Dashboard', icon: FiGrid },
  { to: '/matches', label: 'Find Collaborators', icon: FiUsers },
  { to: '/my-profile', label: 'My Profile', icon: FiUser },
  { to: '/assessment', label: 'Skill Assessment', icon: FiAward },
  { to: '/leaderboard', label: 'Leaderboard', icon: FiBarChart2 },
];

const Layout: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { user, logout } = useAuth();
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const [pendingCount, setPendingCount] = useState(0);
  const location = useLocation();

  const fetchPendingCount = useCallback(async () => {
    if (!user) return;
    try {
      const response: APIResponse<Match[]> = await api.get('/matches');
      if (response.success && response.data) {
        const incoming = response.data.filter(
          (m) => m.user2.id === user.id && m.status === 'pending'
        );
        setPendingCount(incoming.length);
      }
    } catch {
      // silently fail
    }
  }, [user]);

  useEffect(() => {
    fetchPendingCount();
  }, [fetchPendingCount, location.pathname]);

  const avatarUrl = `https://ui-avatars.com/api/?name=${encodeURIComponent(
    user?.full_name || user?.username || 'U'
  )}&background=6D28D9&color=fff&bold=true&rounded=true`;

  const SidebarContent = () => (
    <div className="flex h-full flex-col bg-white border-r border-border">
      {/* Logo */}
      <div className="flex h-20 items-center justify-center border-b border-border">
        <Link to="/dashboard" className="flex items-center gap-2.5">
          <img src="/assets/logo-icon.png" alt="SkillSync" className="h-8 w-8 object-contain" />
          <span className="text-xl font-bold text-text-primary">SkillSync</span>
        </Link>
      </div>

      {/* Nav */}
      <nav className="flex-1 space-y-2 p-4">
        {navItems.map(({ to, label, icon: Icon }) => (
          <NavLink
            key={to}
            to={to}
            onClick={() => setSidebarOpen(false)}
            className={({ isActive }) =>
              `group flex items-center gap-3 rounded-lg px-4 py-3 text-sm font-medium transition-all duration-200 ${
                isActive
                  ? 'bg-primary text-white shadow-md'
                  : 'text-text-secondary hover:bg-gray-100 hover:text-text-primary'
              }`
            }
          >
            <Icon className="h-5 w-5" />
            <span className="flex-1">{label}</span>
            {to === '/matches' && pendingCount > 0 && (
              <span className="ml-auto rounded-full bg-red-500 px-2 py-0.5 text-xs font-bold text-white">
                {pendingCount}
              </span>
            )}
          </NavLink>
        ))}
      </nav>

      {/* User section */}
      <div className="border-t border-border p-4">
        <div className="flex items-center gap-3">
          <img src={avatarUrl} alt="User Avatar" className="h-10 w-10 rounded-full" />
          <div className="flex-1 overflow-hidden">
            <p className="truncate font-semibold text-text-primary">{user?.full_name || user?.username}</p>
            <p className="truncate text-xs text-text-secondary">{user?.email}</p>
          </div>
        </div>
        <button
          onClick={logout}
          className="mt-4 flex w-full items-center gap-3 rounded-lg px-4 py-3 text-sm font-medium text-text-secondary transition-colors hover:bg-gray-100 hover:text-text-primary"
        >
          <FiLogOut className="h-5 w-5" />
          Sign Out
        </button>
      </div>
    </div>
  );

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Mobile Sidebar */}
      <Transition.Root show={sidebarOpen} as={Fragment}>
        <div className="fixed inset-0 flex z-40 lg:hidden">
          <Transition.Child
            as={Fragment}
            enter="transition-opacity ease-linear duration-300"
            enterFrom="opacity-0"
            enterTo="opacity-100"
            leave="transition-opacity ease-linear duration-300"
            leaveFrom="opacity-100"
            leaveTo="opacity-0"
          >
            <div className="fixed inset-0 bg-gray-600 bg-opacity-75" onClick={() => setSidebarOpen(false)} />
          </Transition.Child>
          <Transition.Child
            as={Fragment}
            enter="transition ease-in-out duration-300 transform"
            enterFrom="-translate-x-full"
            enterTo="translate-x-0"
            leave="transition ease-in-out duration-300 transform"
            leaveFrom="translate-x-0"
            leaveTo="-translate-x-full"
          >
            <div className="relative flex-1 flex flex-col max-w-xs w-full">
              <SidebarContent />
            </div>
          </Transition.Child>
          <div className="flex-shrink-0 w-14" aria-hidden="true"></div>
        </div>
      </Transition.Root>

      {/* Static Sidebar for desktop */}
      <div className="hidden lg:flex lg:flex-col lg:w-64 lg:fixed lg:inset-y-0 lg:z-30">
        <SidebarContent />
      </div>

      {/* Main content */}
      <div className="lg:pl-64 flex flex-col flex-1">
        <div className="sticky top-0 z-10 lg:hidden pl-3 pt-3 sm:pl-5 sm:pt-5 bg-gray-50">
            <button
            type="button"
            className="-ml-0.5 -mt-0.5 h-12 w-12 inline-flex items-center justify-center rounded-md text-gray-500 hover:text-gray-900 focus:outline-none focus:ring-2 focus:ring-inset focus:ring-indigo-500"
            onClick={() => setSidebarOpen(true)}
            >
            <span className="sr-only">Open sidebar</span>
            <FiMenu className="h-6 w-6" aria-hidden="true" />
            </button>
        </div>
        <main className="flex-1">
          <div className="py-8 px-4 sm:px-6 lg:px-8 max-w-7xl mx-auto">
            {children}
          </div>
        </main>
      </div>
    </div>
  );
};

export default Layout;
