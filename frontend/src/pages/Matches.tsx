import React, { useState, useEffect, useCallback } from 'react';
import { Link } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import api from '../services/api';
import { Match, MatchRequest, APIResponse } from '../types';
import toast from 'react-hot-toast';
import { Loader2, Handshake, MessageSquare, Sparkles, UserCheck, CheckCircle, XCircle, ArrowDownLeft, ArrowUpRight } from 'lucide-react';

type Tab = 'incoming' | 'outgoing' | 'active';

const tabs: { key: Tab; label: string; icon: React.ReactNode }[] = [
  { key: 'incoming', label: 'Incoming Requests', icon: <ArrowDownLeft className="h-4 w-4" /> },
  { key: 'outgoing', label: 'Outgoing Requests', icon: <ArrowUpRight className="h-4 w-4" /> },
  { key: 'active', label: 'Active Matches', icon: <Handshake className="h-4 w-4" /> },
];

const Matches: React.FC = () => {
  const { user } = useAuth();
  const [activeMatches, setActiveMatches] = useState<Match[]>([]);
  const [incomingRequests, setIncomingRequests] = useState<MatchRequest[]>([]);
  const [outgoingRequests, setOutgoingRequests] = useState<MatchRequest[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState<Tab>('incoming');
  const [updatingId, setUpdatingId] = useState<number | null>(null);

  const fetchData = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const [matchesRes, pendingRes] = await Promise.all([
        api.get<{ matches: Match[]; total: number }>('/matches'),
        api.get<{ received: MatchRequest[]; sent: MatchRequest[] }>('/matches/requests/pending'),
      ]);

      if (matchesRes.success && matchesRes.data) {
        setActiveMatches(matchesRes.data.matches || []);
      }
      if (pendingRes.success && pendingRes.data) {
        setIncomingRequests(pendingRes.data.received || []);
        setOutgoingRequests(pendingRes.data.sent || []);
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
    fetchData();
  }, [fetchData]);

  const handleAccept = async (requestId: number) => {
    setUpdatingId(requestId);
    try {
      const response = await api.put<any>(`/matches/request/${requestId}/accept`);
      if (response.success) {
        toast.success('Request accepted!');
        fetchData();
      } else {
        throw new Error(response.error?.message || 'Failed to accept request.');
      }
    } catch (err: any) {
      toast.error(err.message || 'Failed to accept request.');
    } finally {
      setUpdatingId(null);
    }
  };

  const handleReject = async (requestId: number) => {
    setUpdatingId(requestId);
    try {
      const response = await api.put<any>(`/matches/request/${requestId}/reject`);
      if (response.success) {
        toast.success('Request rejected.');
        fetchData();
      } else {
        throw new Error(response.error?.message || 'Failed to reject request.');
      }
    } catch (err: any) {
      toast.error(err.message || 'Failed to reject request.');
    } finally {
      setUpdatingId(null);
    }
  };

  const tabCounts: Record<Tab, number> = {
    incoming: incomingRequests.length,
    outgoing: outgoingRequests.length,
    active: activeMatches.length,
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

      {/* Incoming Requests */}
      {activeTab === 'incoming' && (
        incomingRequests.length === 0 ? (
          <EmptyState message="No incoming requests" detail="Browse collaborators to find your match." />
        ) : (
          <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
            {incomingRequests.map((req) => {
              const sender = req.sender;
              const avatarUrl = `https://ui-avatars.com/api/?name=${encodeURIComponent(sender?.username || 'U')}&background=6366f1&color=fff`;
              const isUpdating = updatingId === req.id;

              return (
                <div key={req.id} className="group rounded-2xl border border-gray-100 bg-white p-5 shadow-sm transition-all duration-300 hover:-translate-y-1 hover:shadow-lg">
                  <div className="flex items-start gap-4">
                    <img src={avatarUrl} alt={sender?.username} className="h-12 w-12 rounded-full ring-2 ring-purple-100" />
                    <div className="flex-1 overflow-hidden">
                      <Link to={`/profile/${sender?.id}`} className="font-semibold text-gray-900 transition-colors hover:text-purple-600">
                        {sender?.full_name || sender?.username}
                      </Link>
                      <p className="text-xs text-gray-400">@{sender?.username}</p>
                    </div>
                    <span className="rounded-full border border-amber-200 bg-amber-50 px-2.5 py-0.5 text-xs font-medium text-amber-600">
                      pending
                    </span>
                  </div>

                  {req.message && (
                    <p className="mt-3 text-sm text-gray-500 italic">"{req.message}"</p>
                  )}

                  <div className="mt-4 flex gap-2">
                    <button
                      onClick={() => handleAccept(req.id)}
                      disabled={isUpdating}
                      className="flex flex-1 items-center justify-center gap-1.5 rounded-xl bg-gradient-to-r from-emerald-500 to-green-500 py-2 text-xs font-medium text-white transition-all hover:scale-[1.02] hover:shadow-lg disabled:opacity-50"
                    >
                      {isUpdating ? <Loader2 className="h-3.5 w-3.5 animate-spin" /> : <CheckCircle className="h-3.5 w-3.5" />}
                      Accept
                    </button>
                    <button
                      onClick={() => handleReject(req.id)}
                      disabled={isUpdating}
                      className="flex flex-1 items-center justify-center gap-1.5 rounded-xl border border-red-200 bg-white py-2 text-xs font-medium text-red-600 transition-all hover:bg-red-50 disabled:opacity-50"
                    >
                      {isUpdating ? <Loader2 className="h-3.5 w-3.5 animate-spin" /> : <XCircle className="h-3.5 w-3.5" />}
                      Reject
                    </button>
                  </div>
                </div>
              );
            })}
          </div>
        )
      )}

      {/* Outgoing Requests */}
      {activeTab === 'outgoing' && (
        outgoingRequests.length === 0 ? (
          <EmptyState message="No outgoing requests" detail="Browse collaborators to send connection requests." />
        ) : (
          <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
            {outgoingRequests.map((req) => {
              const receiver = req.receiver;
              const avatarUrl = `https://ui-avatars.com/api/?name=${encodeURIComponent(receiver?.username || 'U')}&background=6366f1&color=fff`;

              return (
                <div key={req.id} className="group rounded-2xl border border-gray-100 bg-white p-5 shadow-sm transition-all duration-300 hover:-translate-y-1 hover:shadow-lg">
                  <div className="flex items-start gap-4">
                    <img src={avatarUrl} alt={receiver?.username} className="h-12 w-12 rounded-full ring-2 ring-purple-100" />
                    <div className="flex-1 overflow-hidden">
                      <Link to={`/profile/${receiver?.id}`} className="font-semibold text-gray-900 transition-colors hover:text-purple-600">
                        {receiver?.full_name || receiver?.username}
                      </Link>
                      <p className="text-xs text-gray-400">@{receiver?.username}</p>
                    </div>
                    <span className="rounded-full border border-amber-200 bg-amber-50 px-2.5 py-0.5 text-xs font-medium text-amber-600">
                      pending
                    </span>
                  </div>
                  {req.message && (
                    <p className="mt-3 text-sm text-gray-500 italic">"{req.message}"</p>
                  )}
                  <p className="mt-3 text-xs text-gray-400">Waiting for response...</p>
                </div>
              );
            })}
          </div>
        )
      )}

      {/* Active Matches */}
      {activeTab === 'active' && (
        activeMatches.length === 0 ? (
          <EmptyState message="No active matches yet" detail="Accept incoming requests or send new ones to get started." />
        ) : (
          <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
            {activeMatches.map((match) => {
              const partner = match.user1.id === user?.id ? match.user2 : match.user1;
              const avatarUrl = `https://ui-avatars.com/api/?name=${encodeURIComponent(partner.username)}&background=6366f1&color=fff`;

              return (
                <div key={match.id} className="group rounded-2xl border border-gray-100 bg-white p-5 shadow-sm transition-all duration-300 hover:-translate-y-1 hover:shadow-lg">
                  <div className="flex items-start gap-4">
                    <img src={avatarUrl} alt={partner.username} className="h-12 w-12 rounded-full ring-2 ring-purple-100" />
                    <div className="flex-1 overflow-hidden">
                      <Link to={`/profile/${partner.id}`} className="font-semibold text-gray-900 transition-colors hover:text-purple-600">
                        {partner.full_name || partner.username}
                      </Link>
                      <p className="text-xs text-gray-400">@{partner.username}</p>
                    </div>
                    <span className="rounded-full border border-emerald-200 bg-emerald-50 px-2.5 py-0.5 text-xs font-medium text-emerald-600">
                      active
                    </span>
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

                  <div className="mt-4 flex gap-2">
                    <Link
                      to={`/chat/${match.id}`}
                      className="flex flex-1 items-center justify-center gap-1.5 rounded-xl bg-gradient-to-r from-[#6D28D9] via-[#7C3AED] to-[#EC4899] py-2 text-xs font-medium text-white transition-all hover:scale-[1.02] hover:shadow-lg hover:shadow-purple-500/20"
                    >
                      <MessageSquare className="h-3.5 w-3.5" />
                      Chat
                    </Link>
                    <Link
                      to={`/match/${match.id}`}
                      className="flex flex-1 items-center justify-center gap-1.5 rounded-xl border border-gray-200 bg-white py-2 text-xs font-medium text-gray-600 transition-all hover:bg-gray-50 hover:text-gray-900"
                    >
                      <UserCheck className="h-3.5 w-3.5" />
                      Details
                    </Link>
                  </div>
                </div>
              );
            })}
          </div>
        )
      )}
    </div>
  );
};

const EmptyState: React.FC<{ message: string; detail: string }> = ({ message, detail }) => (
  <div className="flex flex-col items-center justify-center rounded-2xl border border-gray-100 bg-white py-16 shadow-sm">
    <Handshake className="h-12 w-12 text-gray-300" />
    <p className="mt-4 text-lg font-medium text-gray-500">{message}</p>
    <p className="mt-1 text-sm text-gray-400">{detail}</p>
    <Link
      to="/users"
      className="mt-6 rounded-xl bg-gradient-to-r from-[#6D28D9] via-[#7C3AED] to-[#EC4899] px-6 py-2.5 text-sm font-medium text-white shadow-lg shadow-purple-500/20 transition-all hover:scale-105"
    >
      Find Collaborators
    </Link>
  </div>
);

export default Matches;
