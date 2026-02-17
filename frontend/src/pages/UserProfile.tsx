import React, { useState, useEffect, useCallback } from 'react';
import { useParams, Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import api from '../services/api';
import { APIResponse, UserProfile as UserProfileType, UserSkill, Rating } from '../types';
import toast from 'react-hot-toast';
import { Loader2, Star, Trophy, CheckCircle, MessageSquare } from 'lucide-react';

const ProgressBar: React.FC<{ label: string; score: number }> = ({ label, score }) => {
  const s = Math.max(0, Math.min(100, score));
  const color = s >= 80 ? 'from-green-400 to-cyan-400' : s >= 60 ? 'from-yellow-400 to-orange-400' : 'from-pink-500 to-red-500';

  return (
    <div className="mb-3">
      <div className="mb-1 flex justify-between">
        <span className="text-xs font-medium text-text-secondary">{label}</span>
        <span className="text-xs font-medium text-text-secondary">{s}/100</span>
      </div>
      <div className="h-1.5 overflow-hidden rounded-full bg-gray-200">
        <div className={`h-full rounded-full bg-gradient-to-r ${color} transition-all duration-500`} style={{ width: `${s}%` }} />
      </div>
    </div>
  );
};

const StarRating: React.FC<{ rating: number }> = ({ rating }) => (
  <div className="flex gap-0.5">
    {[1, 2, 3, 4, 5].map((i) => (
      <Star key={i} className={`h-4 w-4 ${i <= rating ? 'fill-yellow-400 text-yellow-400' : 'text-gray-300'}`} />
    ))}
  </div>
);

const UserProfile: React.FC = () => {
  const { userId } = useParams<{ userId: string }>();
  const { user: currentUser, isAuthenticated } = useAuth();
  const navigate = useNavigate();
  const [profile, setProfile] = useState<UserProfileType | null>(null);
  const [loading, setLoading] = useState(true);
  const [connecting, setConnecting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchUserProfile = useCallback(async () => {
    if (!userId) { setError('User ID is missing.'); setLoading(false); return; }
    setLoading(true);
    setError(null);
    try {
      const response: APIResponse<UserProfileType> = await api.get(`/users/${userId}`);
      if (response.success && response.data) {
        setProfile(response.data);
      } else {
        throw new Error(response.error?.message || 'Failed to fetch user profile.');
      }
    } catch (err: any) {
      setError(err.message || 'An error occurred.');
      toast.error(err.message || 'Failed to load profile.');
    } finally {
      setLoading(false);
    }
  }, [userId]);

  useEffect(() => { fetchUserProfile(); }, [fetchUserProfile]);

  const handleConnect = async () => {
    if (!userId || !currentUser) return;
    setConnecting(true);
    try {
      const mySkills = currentUser.skills_teach || [];
      const theirSkills = profile?.skills_learn || [];
      const response: APIResponse<any> = await api.post('/matches', {
        target_user_id: userId,
        skill_offered: mySkills[0] || 'general',
        skill_wanted: theirSkills[0] || 'general',
      });
      if (response.success) {
        toast.success('Connection request sent!');
        navigate('/matches');
      } else {
        throw new Error(response.error?.message || 'Failed to send connection request.');
      }
    } catch (err: any) {
      toast.error(err.message || 'Failed to connect.');
    } finally {
      setConnecting(false);
    }
  };

  if (loading) {
    return (
      <div className="flex flex-1 items-center justify-center p-8">
        <Loader2 className="h-8 w-8 animate-spin text-primary" />
      </div>
    );
  }

  if (error || !profile) {
    return (
      <div className="flex flex-1 items-center justify-center p-8">
        <div className="rounded-2xl border border-red-200 bg-red-50 px-6 py-4 text-sm text-red-700">
          {error || 'Profile not found.'}
        </div>
      </div>
    );
  }

  const isOwnProfile = isAuthenticated && currentUser?.id === userId;
  const avatarUrl = `https://ui-avatars.com/api/?name=${encodeURIComponent(profile.full_name || profile.username)}&background=7c3aed&color=fff&size=128`;

  return (
    <div className="space-y-6">
      {/* Profile header */}
      <div className="rounded-2xl bg-gradient-to-r from-purple-600 to-indigo-600 p-6 shadow-lg">
        <div className="flex flex-col items-center gap-6 md:flex-row md:items-start">
          <img src={avatarUrl} alt={profile.username} className="h-24 w-24 rounded-full ring-4 ring-white/30 shadow-xl" />
          <div className="flex-1 text-center md:text-left">
            <h1 className="text-3xl font-bold text-white">{profile.full_name || profile.username}</h1>
            <p className="mt-1 text-purple-200">@{profile.username}</p>
            <div className="mt-3 flex flex-wrap items-center justify-center gap-4 md:justify-start">
              <div className="flex items-center gap-1.5">
                <Star className="h-4 w-4 text-yellow-400" />
                <span className="text-lg font-bold text-white">{profile.reputation_score || 0}</span>
                <span className="text-xs text-purple-200">reputation</span>
              </div>
              {profile.average_rating !== undefined && profile.average_rating > 0 && (
                <div className="flex items-center gap-2">
                  <StarRating rating={Math.round(profile.average_rating ?? 0)} />
                  <span className="text-sm text-purple-200">({profile.average_rating?.toFixed(1)})</span>
                </div>
              )}
            </div>
            {!isOwnProfile && isAuthenticated && (
              <button
                onClick={handleConnect}
                disabled={connecting}
                className="mt-4 inline-flex items-center gap-2 rounded-xl bg-white px-5 py-2 text-sm font-medium text-purple-700 shadow-lg transition-all hover:scale-105 hover:bg-purple-50 disabled:opacity-50"
              >
                {connecting ? (
                  <Loader2 className="h-4 w-4 animate-spin" />
                ) : (
                  <MessageSquare className="h-4 w-4" />
                )}
                {connecting ? 'Connecting...' : 'Connect'}
              </button>
            )}
          </div>
        </div>

        {/* Badges */}
        {profile.badges && profile.badges.length > 0 && (
          <div className="mt-6 flex flex-wrap gap-2">
            {profile.badges.map((badge, i) => (
              <span key={i} className="inline-flex items-center gap-1.5 rounded-full bg-white/20 px-3 py-1 text-xs font-medium text-white">
                <Trophy className="h-3 w-3 text-yellow-400" />
                {badge}
              </span>
            ))}
          </div>
        )}
      </div>

      <div className="grid grid-cols-1 gap-4 lg:grid-cols-3">
        {/* Reputation breakdown */}
        <div className="rounded-2xl border border-border bg-card-bg p-5 shadow-card">
          <h2 className="mb-4 text-xs font-medium uppercase tracking-wider text-primary">Reputation Breakdown</h2>
          {profile.reputation_breakdown ? (
            <>
              <ProgressBar label="Code Quality" score={profile.reputation_breakdown.code_quality || 0} />
              <ProgressBar label="Communication" score={profile.reputation_breakdown.communication || 0} />
              <ProgressBar label="Helpfulness" score={profile.reputation_breakdown.helpfulness || 0} />
              <ProgressBar label="Reliability" score={profile.reputation_breakdown.reliability || 0} />
            </>
          ) : (
            <p className="text-sm text-text-secondary">Not available yet.</p>
          )}
        </div>

        {/* Skills */}
        <div className="rounded-2xl border border-border bg-card-bg p-5 shadow-card">
          <h2 className="mb-4 text-xs font-medium uppercase tracking-wider text-primary">Skills</h2>
          {profile.skills && profile.skills.length > 0 ? (
            <ul className="space-y-3">
              {profile.skills.map((skill: UserSkill) => (
                <li key={skill.skill_id} className="flex items-center justify-between">
                  <span className="text-sm font-medium text-text-primary">{skill.skill_id}</span>
                  <div className="flex items-center gap-2">
                    <span className="text-xs text-text-secondary">{skill.credibility_score}%</span>
                    {skill.verified_by_peers && (
                      <span className="inline-flex items-center gap-1 rounded-full bg-green-100 px-2 py-0.5 text-[10px] font-medium text-green-700">
                        <CheckCircle className="h-2.5 w-2.5" />
                        Verified
                      </span>
                    )}
                  </div>
                </li>
              ))}
            </ul>
          ) : (
            <p className="text-sm text-text-secondary">No skills listed yet.</p>
          )}
        </div>

        {/* Stats */}
        <div className="rounded-2xl border border-border bg-card-bg p-5 shadow-card">
          <h2 className="mb-4 text-xs font-medium uppercase tracking-wider text-primary">Statistics</h2>
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <span className="text-sm text-text-secondary">Total Matches</span>
              <span className="text-lg font-bold text-primary">{profile.total_matches || 0}</span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-sm text-text-secondary">Sessions Completed</span>
              <span className="text-lg font-bold text-green-600">{profile.sessions_completed || 0}</span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-sm text-text-secondary">Success Rate</span>
              <span className="text-lg font-bold text-blue-600">{((profile.success_rate || 0) * 100).toFixed(1)}%</span>
            </div>
          </div>
        </div>
      </div>

      {/* Recent ratings */}
      {profile.recent_ratings && profile.recent_ratings.length > 0 && (
        <div className="rounded-2xl border border-border bg-card-bg p-5 shadow-card">
          <h2 className="mb-4 text-xs font-medium uppercase tracking-wider text-primary">Recent Ratings</h2>
          <div className="grid grid-cols-1 gap-3 md:grid-cols-2">
            {profile.recent_ratings.map((rating: Rating) => (
              <div key={rating.id} className="rounded-xl border border-border bg-gray-50 p-4">
                <div className="flex items-center justify-between">
                  <span className="text-xs text-text-secondary">Rated by: {rating.rater_id}</span>
                  <StarRating rating={rating.overall_rating} />
                </div>
                {rating.feedback && (
                  <p className="mt-2 text-sm italic text-text-secondary">"{rating.feedback}"</p>
                )}
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
};

export default UserProfile;
