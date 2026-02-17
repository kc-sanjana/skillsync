import React, { useState, useEffect, useCallback } from 'react';
import { Link } from 'react-router-dom';
import api from '../services/api';
import { APIResponse, PaginatedResponse, User } from '../types';
import { useAuth } from '../contexts/AuthContext';
import toast from 'react-hot-toast';
import { FiLoader, FiAward, FiChevronDown, FiBarChart2 } from 'react-icons/fi';
import { Trophy as FiTrophy } from 'lucide-react';

type LeaderboardCategory = 'overall' | 'code_quality' | 'communication' | 'helpfulness' | 'reliability';

interface LeaderboardUser extends User {
  rank: number;
  score: number;
}

const categories: { key: LeaderboardCategory; label: string }[] = [
  { key: 'overall', label: 'Overall Reputation' },
  { key: 'code_quality', label: 'Code Quality' },
  { key: 'communication', label: 'Communication' },
  { key: 'helpfulness', label: 'Helpfulness' },
  { key: 'reliability', label: 'Reliability' },
];

const Leaderboard: React.FC = () => {
  const { user: currentUser } = useAuth();
  const [activeCategory, setActiveCategory] = useState<LeaderboardCategory>('overall');
  const [leaderboardUsers, setLeaderboardUsers] = useState<LeaderboardUser[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchLeaderboard = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const response: APIResponse<PaginatedResponse<LeaderboardUser>> = await api.get('/leaderboard', {
        params: {
          category: activeCategory,
          limit: 10,
        }
      });

      if (response.success && response.data) {
        const items = response.data.data || [];
        const ranked = items.map((u: any, i: number) => ({
          ...u,
          id: u.id || u.user_id,
          score: u.score ?? u.overall_score ?? 0,
          full_name: u.full_name || u.username || '',
          reputation_score: u.reputation_score ?? u.overall_score ?? 0,
          rank: u.rank || i + 1,
        }));
        setLeaderboardUsers(ranked);
      } else {
        setLeaderboardUsers([]);
      }
    } catch (err: any) {
      setError(err.message || 'An error occurred.');
      toast.error(err.message || 'Failed to load leaderboard.');
    } finally {
      setLoading(false);
    }
  }, [activeCategory]);

  useEffect(() => { fetchLeaderboard(); }, [fetchLeaderboard]);
  
  const getRankClasses = (rank: number) => {
    switch (rank) {
      case 1: return 'border-yellow-400 bg-yellow-400/10';
      case 2: return 'border-gray-400 bg-gray-400/10';
      case 3: return 'border-orange-400 bg-orange-400/10';
      default: return 'border-border';
    }
  };
  
  const RankIcon: React.FC<{rank: number}> = ({ rank }) => {
    switch (rank) {
      case 1: return <FiAward className="h-6 w-6 text-yellow-500" />;
      case 2: return <FiAward className="h-6 w-6 text-gray-500" />;
      case 3: return <FiAward className="h-6 w-6 text-orange-500" />;
      default: return <span className="text-text-secondary font-semibold">#{rank}</span>;
    }
  };


  return (
    <div className="space-y-8">
      <div className="text-center">
        <FiTrophy className="mx-auto h-12 w-12 text-primary" />
        <h1 className="mt-4 text-4xl font-bold text-text-primary">Leaderboard</h1>
        <p className="mt-2 text-lg text-text-secondary">See who's making the biggest impact in the community.</p>
      </div>

      {/* Filters */}
      <div className="flex items-center justify-center flex-wrap gap-2">
        {categories.map(({ key, label }) => (
            <button
            key={key}
            onClick={() => setActiveCategory(key)}
            className={`rounded-full px-4 py-2 text-sm font-medium transition-all duration-200 ${
                activeCategory === key
                ? 'bg-primary text-white shadow'
                : 'bg-card-bg text-text-secondary hover:bg-gray-100 border border-border'
            }`}
            >
            {label}
            </button>
        ))}
      </div>

      {/* Leaderboard list */}
      <div className="max-w-3xl mx-auto">
        {loading ? (
          <div className="flex items-center justify-center py-20">
            <FiLoader className="h-10 w-10 animate-spin text-primary" />
          </div>
        ) : error ? (
          <div className="rounded-large border border-red-200 bg-red-50 px-6 py-4 text-center text-sm text-red-700">{error}</div>
        ) : leaderboardUsers.length === 0 ? (
            <div className="text-center py-20">
                <h3 className="text-xl font-semibold text-text-primary">No Data Available</h3>
                <p className="mt-1 text-text-secondary">There are no rankings for this category yet.</p>
            </div>
        ) : (
          <div className="space-y-3">
            {leaderboardUsers.map((u) => {
              const isMe = currentUser?.id === u.id;
              const avatarUrl = `https://ui-avatars.com/api/?name=${encodeURIComponent(u.full_name || u.username)}&background=random&color=fff&bold=true&rounded=true`;

              return (
                <div
                  key={u.id}
                  className={`bg-card-bg rounded-large p-4 flex items-center gap-4 border-2 transition-all ${getRankClasses(u.rank)} ${isMe ? 'shadow-lg scale-[1.02]' : 'shadow-card'}`}
                >
                  <div className="w-10 text-center">
                    <RankIcon rank={u.rank} />
                  </div>
                  
                  <img src={avatarUrl} alt={u.username} className="h-12 w-12 rounded-full" />
                  
                  <div className="flex-1 overflow-hidden">
                    <Link to={`/profile/${u.id}`} className="font-semibold text-text-primary hover:text-primary transition-colors truncate block">
                      {u.full_name || u.username}
                      {isMe && <span className="ml-2 text-xs text-primary">(You)</span>}
                    </Link>
                    <p className="text-sm text-text-secondary">@{u.username}</p>
                  </div>
                  
                  <div className="flex items-center gap-2 text-lg font-bold text-text-primary">
                    <FiBarChart2 className="text-primary"/>
                    <span>{(u.score || 0).toFixed(1)}</span>
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </div>
    </div>
  );
};

export default Leaderboard;
