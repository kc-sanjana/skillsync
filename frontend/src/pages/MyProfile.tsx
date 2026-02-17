import React, { useState, useEffect, useCallback } from 'react';
import { Link } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import api from '../services/api';
import { Match, User, APIResponse } from '../types';
import toast from 'react-hot-toast';
import { Loader2, Edit3, Save, X, Star, Trophy, ArrowDownLeft } from 'lucide-react';

const MyProfile: React.FC = () => {
  const { user, refreshUser } = useAuth();
  const [editing, setEditing] = useState(false);
  const [saving, setSaving] = useState(false);
  const [pendingCount, setPendingCount] = useState(0);
  const [loadingMatches, setLoadingMatches] = useState(true);

  const [form, setForm] = useState({
    full_name: '',
    bio: '',
    skills_teach: '',
    skills_learn: '',
  });

  useEffect(() => {
    if (user) {
      setForm({
        full_name: user.full_name || '',
        bio: user.bio || '',
        skills_teach: (user.skills_teach || []).join(', '),
        skills_learn: (user.skills_learn || []).join(', '),
      });
    }
  }, [user]);

  const fetchPendingCount = useCallback(async () => {
    setLoadingMatches(true);
    try {
      const response: APIResponse<Match[]> = await api.get('/matches');
      if (response.success && response.data) {
        const incoming = response.data.filter(
          (m) => m.user2.id === user?.id && m.status === 'pending'
        );
        setPendingCount(incoming.length);
      }
    } catch {
      // silently fail
    } finally {
      setLoadingMatches(false);
    }
  }, [user?.id]);

  useEffect(() => {
    if (user) fetchPendingCount();
  }, [user, fetchPendingCount]);

  const handleSave = async () => {
    setSaving(true);
    try {
      const payload = {
        full_name: form.full_name.trim(),
        bio: form.bio.trim(),
        skills_teach: form.skills_teach.split(',').map((s) => s.trim()).filter(Boolean),
        skills_learn: form.skills_learn.split(',').map((s) => s.trim()).filter(Boolean),
      };
      const response: APIResponse<User> = await api.put('/users/me', payload);
      if (response.success) {
        toast.success('Profile updated!');
        await refreshUser();
        setEditing(false);
      } else {
        throw new Error(response.error?.message || 'Failed to update profile.');
      }
    } catch (err: any) {
      toast.error(err.message || 'Failed to update profile.');
    } finally {
      setSaving(false);
    }
  };

  const handleCancel = () => {
    if (user) {
      setForm({
        full_name: user.full_name || '',
        bio: user.bio || '',
        skills_teach: (user.skills_teach || []).join(', '),
        skills_learn: (user.skills_learn || []).join(', '),
      });
    }
    setEditing(false);
  };

  if (!user) {
    return (
      <div className="flex flex-1 items-center justify-center p-8">
        <Loader2 className="h-8 w-8 animate-spin text-primary" />
      </div>
    );
  }

  const avatarUrl = `https://ui-avatars.com/api/?name=${encodeURIComponent(
    user.full_name || user.username
  )}&background=7c3aed&color=fff&size=128`;

  return (
    <div className="space-y-6">
      {/* Profile Header */}
      <div className="rounded-2xl bg-gradient-to-r from-purple-600 to-indigo-600 p-6 shadow-lg">
        <div className="flex flex-col items-center gap-6 md:flex-row md:items-start">
          <img
            src={avatarUrl}
            alt={user.username}
            className="h-24 w-24 rounded-full ring-4 ring-white/30 shadow-xl"
          />
          <div className="flex-1 text-center md:text-left">
            <h1 className="text-3xl font-bold text-white">
              {user.full_name || user.username}
            </h1>
            <p className="mt-1 text-purple-200">@{user.username}</p>
            <p className="mt-1 text-sm text-purple-200">{user.email}</p>
            <div className="mt-3 flex flex-wrap items-center justify-center gap-4 md:justify-start">
              <div className="flex items-center gap-1.5">
                <Star className="h-4 w-4 text-yellow-400" />
                <span className="text-lg font-bold text-white">{user.reputation_score || 0}</span>
                <span className="text-xs text-purple-200">reputation</span>
              </div>
              <div className="flex items-center gap-1.5">
                <Trophy className="h-4 w-4 text-yellow-400" />
                <span className="text-lg font-bold text-white">{user.total_sessions || 0}</span>
                <span className="text-xs text-purple-200">sessions</span>
              </div>
            </div>
          </div>
          {!editing && (
            <button
              onClick={() => setEditing(true)}
              className="inline-flex items-center gap-2 rounded-xl bg-white px-5 py-2 text-sm font-medium text-purple-700 shadow-lg transition-all hover:scale-105 hover:bg-purple-50"
            >
              <Edit3 className="h-4 w-4" />
              Edit Profile
            </button>
          )}
        </div>

        {/* Badges */}
        {user.badges && user.badges.length > 0 && (
          <div className="mt-6 flex flex-wrap gap-2">
            {user.badges.map((badge, i) => (
              <span
                key={i}
                className="inline-flex items-center gap-1.5 rounded-full bg-white/20 px-3 py-1 text-xs font-medium text-white"
              >
                <Trophy className="h-3 w-3 text-yellow-400" />
                {badge}
              </span>
            ))}
          </div>
        )}
      </div>

      {/* Pending Requests Banner */}
      {!loadingMatches && pendingCount > 0 && (
        <Link
          to="/matches"
          className="flex items-center gap-3 rounded-2xl border border-amber-200 bg-amber-50 p-4 transition-all hover:bg-amber-100"
        >
          <div className="rounded-lg bg-amber-100 p-2">
            <ArrowDownLeft className="h-5 w-5 text-amber-600" />
          </div>
          <div className="flex-1">
            <p className="text-sm font-semibold text-amber-800">
              {pendingCount} pending connection {pendingCount === 1 ? 'request' : 'requests'}
            </p>
            <p className="text-xs text-amber-600">Click to view and respond</p>
          </div>
        </Link>
      )}

      {/* Profile Details / Edit Form */}
      <div className="rounded-2xl border border-border bg-card-bg p-6 shadow-card">
        {editing ? (
          <div className="space-y-5">
            <h2 className="text-lg font-bold text-text-primary">Edit Profile</h2>

            <div>
              <label className="mb-1.5 block text-sm font-medium text-text-secondary">Full Name</label>
              <input
                type="text"
                value={form.full_name}
                onChange={(e) => setForm({ ...form, full_name: e.target.value })}
                className="w-full rounded-lg border border-gray-300 px-4 py-2.5 text-sm text-text-primary focus:border-purple-500 focus:outline-none focus:ring-2 focus:ring-purple-500/20"
              />
            </div>

            <div>
              <label className="mb-1.5 block text-sm font-medium text-text-secondary">Bio</label>
              <textarea
                value={form.bio}
                onChange={(e) => setForm({ ...form, bio: e.target.value })}
                rows={3}
                className="w-full rounded-lg border border-gray-300 px-4 py-2.5 text-sm text-text-primary focus:border-purple-500 focus:outline-none focus:ring-2 focus:ring-purple-500/20"
                placeholder="Tell others about yourself..."
              />
            </div>

            <div>
              <label className="mb-1.5 block text-sm font-medium text-text-secondary">Skills I Can Teach</label>
              <input
                type="text"
                value={form.skills_teach}
                onChange={(e) => setForm({ ...form, skills_teach: e.target.value })}
                className="w-full rounded-lg border border-gray-300 px-4 py-2.5 text-sm text-text-primary focus:border-purple-500 focus:outline-none focus:ring-2 focus:ring-purple-500/20"
                placeholder="e.g. React, Python, Design (comma separated)"
              />
            </div>

            <div>
              <label className="mb-1.5 block text-sm font-medium text-text-secondary">Skills I Want to Learn</label>
              <input
                type="text"
                value={form.skills_learn}
                onChange={(e) => setForm({ ...form, skills_learn: e.target.value })}
                className="w-full rounded-lg border border-gray-300 px-4 py-2.5 text-sm text-text-primary focus:border-purple-500 focus:outline-none focus:ring-2 focus:ring-purple-500/20"
                placeholder="e.g. Go, Machine Learning, UX (comma separated)"
              />
            </div>

            <div className="flex gap-3">
              <button
                onClick={handleSave}
                disabled={saving}
                className="inline-flex items-center gap-2 rounded-xl bg-gradient-to-r from-[#6D28D9] via-[#7C3AED] to-[#EC4899] px-6 py-2.5 text-sm font-medium text-white shadow-lg shadow-purple-500/20 transition-all hover:scale-105 disabled:opacity-50"
              >
                {saving ? <Loader2 className="h-4 w-4 animate-spin" /> : <Save className="h-4 w-4" />}
                {saving ? 'Saving...' : 'Save Changes'}
              </button>
              <button
                onClick={handleCancel}
                disabled={saving}
                className="inline-flex items-center gap-2 rounded-xl border border-gray-200 bg-white px-6 py-2.5 text-sm font-medium text-gray-600 transition-all hover:bg-gray-50 disabled:opacity-50"
              >
                <X className="h-4 w-4" />
                Cancel
              </button>
            </div>
          </div>
        ) : (
          <div className="space-y-6">
            <div>
              <h2 className="text-xs font-medium uppercase tracking-wider text-primary">About</h2>
              <p className="mt-2 text-sm text-text-primary">
                {user.bio || 'No bio added yet.'}
              </p>
            </div>

            <div>
              <h2 className="text-xs font-medium uppercase tracking-wider text-primary">Skills I Can Teach</h2>
              <div className="mt-2 flex flex-wrap gap-2">
                {user.skills_teach && user.skills_teach.length > 0 ? (
                  user.skills_teach.map((skill) => (
                    <span key={skill} className="rounded-full bg-purple-50 px-3 py-1 text-sm text-purple-700">
                      {skill}
                    </span>
                  ))
                ) : (
                  <p className="text-sm text-text-secondary">No teaching skills listed yet.</p>
                )}
              </div>
            </div>

            <div>
              <h2 className="text-xs font-medium uppercase tracking-wider text-primary">Skills I Want to Learn</h2>
              <div className="mt-2 flex flex-wrap gap-2">
                {user.skills_learn && user.skills_learn.length > 0 ? (
                  user.skills_learn.map((skill) => (
                    <span key={skill} className="rounded-full bg-pink-50 px-3 py-1 text-sm text-pink-700">
                      {skill}
                    </span>
                  ))
                ) : (
                  <p className="text-sm text-text-secondary">No learning skills listed yet.</p>
                )}
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

export default MyProfile;
