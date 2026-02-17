import React, { useState, useEffect, useCallback } from 'react';
import { Link } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import api from '../services/api';
import { Match, APIResponse } from '../types';
import { FiUsers, FiClipboard, FiStar, FiArrowRight, FiSearch, FiAward, FiTrendingUp, FiLoader, FiInbox } from 'react-icons/fi';

interface DashboardStats {
  totalMatches: number;
  completedSessions: number;
  reputationScore: number;
  pendingRequests: number;
}

const Dashboard: React.FC = () => {
  const { user, isAuthenticated, loading: authLoading } = useAuth();
  const [stats, setStats] = useState<DashboardStats | null>(null);
  const [dataLoading, setDataLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchData = useCallback(async () => {
    if (!isAuthenticated || !user) {
      setDataLoading(false);
      return;
    }

    try {
      setDataLoading(true);
      setError(null);

      const matchesResponse: APIResponse<Match[]> = await api.get('/matches');
      const matchesData = matchesResponse?.success && matchesResponse.data ? matchesResponse.data : [];
      const totalMatches = matchesData.length;
      const pendingRequests = matchesData.filter(
        (m) => m.user2.id === user.id && m.status === 'pending'
      ).length;

      setStats({
        totalMatches,
        completedSessions: user.total_sessions || 0,
        reputationScore: parseFloat(user.reputation_score?.toFixed(1) || '0.0'),
        pendingRequests,
      });
    } catch (err: any) {
      console.error('Failed to fetch dashboard data:', err);
      setError('Failed to load dashboard data.');
    } finally {
      setDataLoading(false);
    }
  }, [isAuthenticated, user]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const statCards = [
    { key: 'totalMatches' as const, label: 'Total Matches', icon: FiUsers, value: stats?.totalMatches, link: undefined as string | undefined },
    { key: 'pendingRequests' as const, label: 'Pending Requests', icon: FiInbox, value: stats?.pendingRequests, link: '/matches' },
    { key: 'completedSessions' as const, label: 'Sessions Completed', icon: FiClipboard, value: stats?.completedSessions, link: undefined as string | undefined },
    { key: 'reputationScore' as const, label: 'Reputation Score', icon: FiStar, value: stats?.reputationScore, link: undefined as string | undefined },
  ];

  const quickActions = [
    { to: '/matches', label: 'Find Collaborators', icon: FiSearch, description: 'Browse and connect with users.' },
    { to: '/assessment', label: 'Take an Assessment', icon: FiAward, description: 'Test and certify your skills.' },
    { to: '/leaderboard', label: 'View Leaderboard', icon: FiTrendingUp, description: 'See who is at the top.' },
  ];

  const avatarUrl = `https://ui-avatars.com/api/?name=${encodeURIComponent(
    user?.full_name || user?.username || 'U'
  )}&background=6D28D9&color=fff&bold=true&rounded=true`;

  if (authLoading || dataLoading) {
    return (
      <div className="flex flex-1 items-center justify-center p-8">
        <FiLoader className="h-8 w-8 animate-spin text-primary" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex flex-1 items-center justify-center p-8">
        <div className="rounded-large border border-red-200 bg-red-50 px-6 py-4 text-sm text-red-700">{error}</div>
      </div>
    );
  }
  
  if (!isAuthenticated || !user) {
    return (
      <div className="flex flex-1 items-center justify-center p-8 text-text-secondary">
        Please log in to view your dashboard.
      </div>
    );
  }

  return (
    <div className="space-y-12">
      {/* Welcome Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div className='flex items-center gap-4'>
          <img src={avatarUrl} alt="User Avatar" className="h-16 w-16 rounded-full" />
          <div>
            <h1 className="text-3xl font-bold text-text-primary">
              Welcome Back, {user.full_name || user.username}!
            </h1>
            <p className="mt-1 text-text-secondary">Here's your SkillSync overview.</p>
          </div>
        </div>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-4">
        {statCards.map(({ key, label, icon: Icon, value, link }) => {
          const card = (
            <div className={`bg-card-bg rounded-large shadow-card p-6 transition-transform duration-300 ease-in-out hover:-translate-y-1 hover:shadow-card-hover ${link ? 'cursor-pointer' : ''}`}>
              <div className="flex items-center justify-between">
                <p className="text-sm font-medium text-text-secondary">{label}</p>
                <div className="bg-primary/10 p-2 rounded-lg">
                  <Icon className="h-5 w-5 text-primary" />
                </div>
              </div>
              <p className="mt-4 text-4xl font-bold text-text-primary">{value ?? '0'}</p>
              {link && (value ?? 0) > 0 && (
                <p className="mt-2 text-xs font-medium text-primary">View requests →</p>
              )}
            </div>
          );
          return link ? (
            <Link key={key} to={link}>{card}</Link>
          ) : (
            <div key={key}>{card}</div>
          );
        })}
      </div>

      {/* Quick Actions */}
      <div>
        <h2 className="text-2xl font-bold text-text-primary mb-4">Quick Actions</h2>
        <div className="grid grid-cols-1 gap-6 md:grid-cols-3">
          {quickActions.map(({ to, label, icon: Icon, description }) => (
            <Link key={to} to={to} className="group block">
                <div className="relative bg-card-bg rounded-large shadow-card p-6 h-full transition-shadow duration-300 ease-in-out group-hover:shadow-card-hover overflow-hidden">
                    <div className="absolute top-0 left-0 h-1 w-full bg-gradient-primary"></div>
                    <div className="flex items-center gap-4">
                        <div className="bg-primary/10 p-3 rounded-lg">
                            <Icon className="h-6 w-6 text-primary" />
                        </div>
                        <div>
                            <p className="font-semibold text-text-primary text-lg">{label}</p>
                        </div>
                    </div>
                    <p className="mt-3 text-sm text-text-secondary">{description}</p>
                </div>
            </Link>
          ))}
        </div>
      </div>
    </div>
  );
};

export default Dashboard;
