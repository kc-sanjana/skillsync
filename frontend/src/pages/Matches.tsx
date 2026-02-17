import React, { useState, useEffect, useCallback } from 'react';
import { Link } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import api from '../services/api';
import { Match, APIResponse } from '../types';
import toast from 'react-hot-toast';
import { Loader2, Handshake, MessageSquare, Sparkles, UserCheck, Clock, CheckCircle, XCircle, ArrowDownLeft, ArrowUpRight } from 'lucide-react';

const statusColors: Record<string, string> = {
  pending: 'border-amber-200 bg-amber-50 text-amber-600',
  accepted: 'border-emerald-200 bg-emerald-50 text-emerald-600',
  rejected: 'border-red-200 bg-red-50 text-red-600',
  completed: 'border-purple-200 bg-purple-50 text-purple-600',
};

type Tab = 'incoming' | 'outgoing' | 'active';

const tabs: { key: Tab; label: string; icon: React.ReactNode }[] = [
  { key: 'incoming', label: 'Incoming Requests', icon: <ArrowDownLeft className="h-4 w-4" /> },
  { key: 'outgoing', label: 'Outgoing Requests', icon: <ArrowUpRight className="h-4 w-4" /> },
  { key: 'active', label: 'Active Matches', icon: <Handshake className="h-4 w-4" /> },
];

const Matches: React.FC = () => {
  const { user } = useAuth();
  const [matches, setMatches] = useState<Match[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState<Tab>('incoming');
  const [updatingId, setUpdatingId] = useState<string | null>(null);

  const fetchMatches = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const response: APIResponse<Match[]> = await api.get('/matches');
      if (response.success && response.data) {
        setMatches(response.data);
      } else {
        throw new Error(response.error?.message || 'Failed to fetch matches.');
      }
    } catch (err: any) {
      console.error('Error fetching matches:', err);
      setError(err.message || 'An error occurred.');
      toast.error(err.message || 'Failed to load matches.');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchMatches();
  }, [fetchMatches]);

  const handleUpdateStatus = async (matchId: string, status: 'accepted' | 'rejected') => {
    setUpdatingId(matchId);
    try {
      const response: APIResponse<Match> = await api.put(`/matches/${matchId}/status`, { status });
      if (response.success) {
        toast.success(status === 'accepted' ? 'Request accepted!' : 'Request rejected.');
        setMatches((prev) =>
          prev.map((m) => (m.id === matchId ? { ...m, status } : m))
        );
      } else {
        throw new Error(response.error?.message || `Failed to ${status} request.`);
      }
    } catch (err: any) {
      toast.error(err.message || `Failed to ${status} request.`);
    } finally {
      setUpdatingId(null);
    }
  };

  const incoming = matches.filter(
    (m) => m.user2.id === user?.id && m.status === 'pending'
  );
  const outgoing = matches.filter(
    (m) => m.user1.id === user?.id && m.status === 'pending'
  );
  const active = matches.filter(
    (m) => m.status === 'accepted' || m.status === 'completed'
  );

  const filteredMatches = activeTab === 'incoming' ? incoming : activeTab === 'outgoing' ? outgoing : active;

  const tabCounts: Record<Tab, number> = {
    incoming: incoming.length,
    outgoing: outgoing.length,
    active: active.length,
  };

  if (loading) {
    return (
      <div className="flex flex-1 items-center justify-center p-8">
        <Loader2 className="h-8 w-8 animate-spin text-purple-500" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex flex-1 items-center justify-center p-8">
        <div className="rounded-2xl border border-red-200 bg-red-50 px-6 py-4 text-sm text-red-600">{error}</div>
      </div>
    );
  }

  return (
    <div className="space-y-6 p-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">Matches</h1>
        <p className="mt-1 text-sm text-gray-400">Your skill exchange pairings</p>
      </div>

      {/* Tabs */}
      <div className="flex gap-1 rounded-xl bg-gray-100 p-1">
        {tabs.map(({ key, label, icon }) => (
          <button
            key={key}
            onClick={() => setActiveTab(key)}
            className={`flex flex-1 items-center justify-center gap-2 rounded-lg px-4 py-2.5 text-sm font-medium transition-all ${
              activeTab === key
                ? 'bg-white text-gray-900 shadow-sm'
                : 'text-gray-500 hover:text-gray-700'
            }`}
          >
            {icon}
            <span className="hidden sm:inline">{label}</span>
            {tabCounts[key] > 0 && (
              <span className={`ml-1 rounded-full px-2 py-0.5 text-xs font-semibold ${
                activeTab === key
                  ? key === 'incoming' ? 'bg-amber-100 text-amber-700' : 'bg-purple-100 text-purple-700'
                  : 'bg-gray-200 text-gray-600'
              }`}>
                {tabCounts[key]}
              </span>
            )}
          </button>
        ))}
      </div>

      {filteredMatches.length === 0 ? (
        <div className="flex flex-col items-center justify-center rounded-2xl border border-gray-100 bg-white py-16 shadow-sm">
          <Handshake className="h-12 w-12 text-gray-300" />
          <p className="mt-4 text-lg font-medium text-gray-500">
            {activeTab === 'incoming' && 'No incoming requests'}
            {activeTab === 'outgoing' && 'No outgoing requests'}
            {activeTab === 'active' && 'No active matches yet'}
          </p>
          <p className="mt-1 text-sm text-gray-400">
            {activeTab === 'active' ? 'Accept incoming requests or send new ones to get started.' : 'Browse collaborators to find your match.'}
          </p>
          <Link
            to="/users"
            className="mt-6 rounded-xl bg-gradient-to-r from-[#6D28D9] via-[#7C3AED] to-[#EC4899] px-6 py-2.5 text-sm font-medium text-white shadow-lg shadow-purple-500/20 transition-all hover:scale-105"
          >
            Find Collaborators
          </Link>
        </div>
      ) : (
        <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
          {filteredMatches.map((match) => {
            const partner = match.user1.id === user?.id ? match.user2 : match.user1;
            const avatarUrl = `https://ui-avatars.com/api/?name=${encodeURIComponent(partner.username)}&background=6366f1&color=fff`;
            const isIncoming = activeTab === 'incoming';
            const isUpdating = updatingId === match.id;

            return (
              <div
                key={match.id}
                className={`group rounded-2xl border border-gray-100 bg-white p-5 shadow-sm transition-all duration-300 hover:-translate-y-1 hover:shadow-lg ${
                  match.status === 'rejected' ? 'opacity-50' : ''
                }`}
              >
                <div className="flex items-start gap-4">
                  <img src={avatarUrl} alt={partner.username} className="h-12 w-12 rounded-full ring-2 ring-purple-100" />
                  <div className="flex-1 overflow-hidden">
                    <Link to={`/profile/${partner.id}`} className="font-semibold text-gray-900 transition-colors hover:text-purple-600">
                      {partner.full_name || partner.username}
                    </Link>
                    <p className="text-xs text-gray-400">@{partner.username}</p>
                  </div>
                  <span className={`rounded-full border px-2.5 py-0.5 text-xs font-medium capitalize ${statusColors[match.status] || 'border-gray-200 bg-gray-50 text-gray-500'}`}>
                    {match.status}
                  </span>
                </div>

                {/* Skills */}
                <div className="mt-4 flex flex-wrap gap-2 text-xs">
                  <span className="rounded-full bg-purple-50 px-2.5 py-1 text-purple-700">{match.skill_offered || 'Skill offered'}</span>
                  <span className="rounded-full bg-pink-50 px-2.5 py-1 text-pink-700">{match.skill_wanted || 'Skill wanted'}</span>
                </div>

                {/* Match score */}
                {match.match_score != null && (
                  <div className="mt-4 flex items-center gap-2">
                    <Sparkles className="h-3.5 w-3.5 text-amber-500" />
                    <div className="flex-1">
                      <div className="h-1.5 overflow-hidden rounded-full bg-gray-100">
                        <div
                          className="h-full rounded-full bg-gradient-to-r from-purple-500 to-pink-500 transition-all duration-500"
                          style={{ width: `${Math.min(100, match.match_score)}%` }}
                        />
                      </div>
                    </div>
                    <span className="text-xs font-medium text-gray-500">{match.match_score}%</span>
                  </div>
                )}

                {/* Actions */}
                <div className="mt-4 flex gap-2">
                  {isIncoming && match.status === 'pending' ? (
                    <>
                      <button
                        onClick={() => handleUpdateStatus(match.id, 'accepted')}
                        disabled={isUpdating}
                        className="flex flex-1 items-center justify-center gap-1.5 rounded-xl bg-gradient-to-r from-emerald-500 to-green-500 py-2 text-xs font-medium text-white transition-all hover:scale-[1.02] hover:shadow-lg disabled:opacity-50"
                      >
                        {isUpdating ? <Loader2 className="h-3.5 w-3.5 animate-spin" /> : <CheckCircle className="h-3.5 w-3.5" />}
                        Accept
                      </button>
                      <button
                        onClick={() => handleUpdateStatus(match.id, 'rejected')}
                        disabled={isUpdating}
                        className="flex flex-1 items-center justify-center gap-1.5 rounded-xl border border-red-200 bg-white py-2 text-xs font-medium text-red-600 transition-all hover:bg-red-50 disabled:opacity-50"
                      >
                        {isUpdating ? <Loader2 className="h-3.5 w-3.5 animate-spin" /> : <XCircle className="h-3.5 w-3.5" />}
                        Reject
                      </button>
                    </>
                  ) : (
                    <>
                      <Link
                        to={`/chat/${match.id}`}
                        className="flex flex-1 items-center justify-center gap-1.5 rounded-xl bg-gradient-to-r from-[#6D28D9] via-[#7C3AED] to-[#EC4899] py-2 text-xs font-medium text-white transition-all hover:scale-[1.02] hover:shadow-lg hover:shadow-purple-500/20"
                      >
                        <MessageSquare className="h-3.5 w-3.5" />
                        Chat
                      </Link>
                      <Link
                        to={`/chat/${match.id}`}
                        className="flex flex-1 items-center justify-center gap-1.5 rounded-xl border border-gray-200 bg-white py-2 text-xs font-medium text-gray-600 transition-all hover:bg-gray-50 hover:text-gray-900"
                      >
                        <UserCheck className="h-3.5 w-3.5" />
                        Details
                      </Link>
                    </>
                  )}
                </div>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
};

export default Matches;
