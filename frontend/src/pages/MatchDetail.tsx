import React, { useState, useEffect, useCallback } from 'react';
import { useParams, Link } from 'react-router-dom';
import api from '../services/api';
import { APIResponse, Match } from '../types';
import { useAuth } from '../contexts/AuthContext';
import toast from 'react-hot-toast';
import { FiLoader, FiMessageSquare, FiCalendar, FiStar, FiZap, FiAward, FiBookOpen, FiShare2 } from 'react-icons/fi';

const MatchGauge: React.FC<{ score: number }> = ({ score }) => {
    const s = Math.max(0, Math.min(100, score));
    const circumference = 2 * Math.PI * 45;
    const offset = circumference - (s / 100) * circumference;
    const color = s >= 80 ? 'text-green-500' : s >= 50 ? 'text-yellow-500' : 'text-red-500';

    return (
        <div className="relative h-32 w-32">
            <svg className="h-full w-full -rotate-90" viewBox="0 0 100 100">
                <circle strokeWidth="8" stroke="currentColor" fill="transparent" r="45" cx="50" cy="50" className="text-gray-200" />
                <circle strokeWidth="8" strokeDasharray={circumference} strokeDashoffset={offset} strokeLinecap="round" stroke="currentColor" fill="transparent" r="45" cx="50" cy="50" className={`${color} transition-all duration-1000 ease-in-out`} />
            </svg>
            <div className="absolute inset-0 flex items-center justify-center">
                <span className={`text-3xl font-bold ${color}`}>{s}%</span>
            </div>
        </div>
    );
};

const MatchDetail: React.FC = () => {
    const { matchId } = useParams<{ matchId: string }>();
    const { user: currentUser } = useAuth();
    const [match, setMatch] = useState<Match | null>(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    const fetchMatchDetails = useCallback(async () => {
        if (!matchId) { setError('Match ID is missing.'); setLoading(false); return; }

        setLoading(true);
        setError(null);
        try {
            const response: APIResponse<Match> = await api.get(`/matches/${matchId}`);
            if (response.success && response.data) {
                setMatch(response.data);
            } else {
                throw new Error(response.error?.message || 'Failed to fetch match details.');
            }
        } catch (err: any) {
            setError(err.message || 'An error occurred.');
            toast.error(err.message || 'Failed to load match details.');
        } finally {
            setLoading(false);
        }
    }, [matchId]);

    useEffect(() => { fetchMatchDetails(); }, [fetchMatchDetails]);

    if (loading) {
        return <div className="flex items-center justify-center p-8"><FiLoader className="h-8 w-8 animate-spin text-primary" /></div>;
    }

    if (error || !match) {
        return <div className="rounded-large border border-red-200 bg-red-50 px-6 py-4 text-sm text-red-700">{error || 'Match not found.'}</div>;
    }

    const partner = match.user1.id === currentUser?.id ? match.user2 : match.user1;
    const avatarUrl = `https://ui-avatars.com/api/?name=${encodeURIComponent(partner.full_name || partner.username)}&background=random&color=fff&size=128`;

    return (
        <div className="space-y-8">
            <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
                <div>
                    <h1 className="text-3xl font-bold text-text-primary">Match with {partner.full_name || partner.username}</h1>
                    <p className="mt-1 text-text-secondary">An overview of your collaboration potential.</p>
                </div>
                <Link to={`/chat/${match.id}`} className="btn btn-primary flex items-center gap-2">
                    <FiMessageSquare className="h-5 w-5"/>
                    <span>Start Chat</span>
                </Link>
            </div>

            <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
                <div className="lg:col-span-1 space-y-8">
                    <div className="bg-card-bg rounded-large shadow-card p-6 text-center">
                        <img src={avatarUrl} alt={partner.username} className="h-24 w-24 rounded-full mx-auto ring-4 ring-primary/20" />
                        <Link to={`/profile/${partner.id}`} className="mt-4 text-xl font-bold text-text-primary hover:text-primary transition-colors">
                            {partner.full_name || partner.username}
                        </Link>
                        <p className="text-sm text-text-secondary">@{partner.username}</p>
                        <div className="mt-2 flex items-center justify-center gap-1.5 text-sm">
                            <FiStar className="h-4 w-4 text-yellow-500" />
                            <span className="font-medium text-text-secondary">{partner.reputation_score?.toFixed(1) || 0} reputation</span>
                        </div>
                    </div>
                     <div className="bg-card-bg rounded-large shadow-card p-6 text-center">
                        <h2 className="text-sm font-medium text-text-secondary mb-4">Match Score</h2>
                        <div className="flex justify-center">
                           <MatchGauge score={match.match_score || 0} />
                        </div>
                        <p className="mt-4 text-xs text-text-secondary">Compatibility based on skills & goals</p>
                    </div>
                </div>

                <div className="lg:col-span-2 bg-card-bg rounded-large shadow-card p-6">
                     <h2 className="text-xl font-bold text-text-primary flex items-center gap-2 mb-6">
                        <FiZap className="text-primary"/>
                        AI Collaboration Insights
                    </h2>
                    {match.ai_insights ? (
                        <div className="space-y-6">
                             <div>
                                <h3 className="font-semibold text-text-primary mb-2 flex items-center gap-2"><FiAward className="text-green-500"/> Skill Complementarity</h3>
                                <ul className="space-y-2 list-disc list-inside text-text-secondary">
                                    {match.ai_insights.skill_complement.map((item, i) => <li key={i}>{item}</li>)}
                                </ul>
                            </div>
                             <div>
                                <h3 className="font-semibold text-text-primary mb-2 flex items-center gap-2"><FiBookOpen className="text-blue-500"/> Learning Opportunities</h3>
                                <ul className="space-y-2 list-disc list-inside text-text-secondary">
                                    {match.ai_insights.learning_opportunities.map((item, i) => <li key={i}>{item}</li>)}
                                </ul>
                            </div>
                            <div>
                                <h3 className="font-semibold text-text-primary mb-2 flex items-center gap-2"><FiShare2 className="text-purple-500"/> Collaboration Ideas</h3>
                                <ul className="space-y-2 list-disc list-inside text-text-secondary">
                                    {match.ai_insights.collaboration_ideas.map((item, i) => <li key={i}>{item}</li>)}
                                </ul>
                            </div>
                            {match.ai_insights.recommendation && (
                                <div className="bg-primary/5 rounded-lg p-4 border border-primary/20">
                                    <p className="text-sm text-text-secondary"><span className="font-semibold text-primary">Recommendation:</span> {match.ai_insights.recommendation}</p>
                                </div>
                            )}
                        </div>
                    ) : (
                        <p className="text-text-secondary">No AI insights are available for this match yet.</p>
                    )}
                </div>
            </div>
        </div>
    );
};

export default MatchDetail;
