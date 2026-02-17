import React from 'react';
import { Star, Trophy, Target, Smile, Award } from 'lucide-react';
import { UserReputation } from '../types';

interface ProgressBarProps {
  label: string;
  score: number;
}

const ProgressBar: React.FC<ProgressBarProps> = ({ label, score }) => {
  const s = Math.max(0, Math.min(100, score));
  const color =
    s >= 80
      ? 'from-emerald-500 to-teal-500'
      : s >= 60
      ? 'from-amber-500 to-orange-500'
      : 'from-red-500 to-pink-500';

  return (
    <div className="mb-3">
      <div className="mb-1 flex justify-between">
        <span className="text-xs font-medium text-gray-600">{label}</span>
        <span className="text-xs font-medium text-gray-400">{s}/100</span>
      </div>
      <div className="h-1.5 overflow-hidden rounded-full bg-gray-100">
        <div
          className={`h-full rounded-full bg-gradient-to-r ${color} transition-all duration-500`}
          style={{ width: `${s}%` }}
        />
      </div>
    </div>
  );
};

const StarRating: React.FC<{ rating: number }> = ({ rating }) => (
  <div className="flex gap-0.5">
    {[1, 2, 3, 4, 5].map((i) => (
      <Star
        key={i}
        className={`h-4 w-4 ${i <= rating ? 'fill-amber-400 text-amber-400' : 'text-gray-200'}`}
      />
    ))}
  </div>
);

const badgeIcons: Record<string, React.ReactNode> = {
  Champion: <Trophy className="h-3.5 w-3.5 text-amber-400" />,
  'Target Achiever': <Target className="h-3.5 w-3.5 text-red-400" />,
  'Team Player': <Smile className="h-3.5 w-3.5 text-blue-400" />,
  'Star Performer': <Award className="h-3.5 w-3.5 text-emerald-400" />,
};

interface ReputationDisplayProps {
  reputation: UserReputation;
  badges: string[];
  averageRating?: number;
  compact?: boolean;
}

const ReputationDisplay: React.FC<ReputationDisplayProps> = ({
  reputation,
  badges,
  averageRating,
  compact = false,
}) => {
  const getOverallColor = (score: number) => {
    if (score >= 90) return 'text-emerald-600';
    if (score >= 70) return 'text-amber-600';
    return 'text-red-600';
  };

  if (compact) {
    return (
      <div className="flex items-center gap-3">
        <span className={`text-xl font-bold ${getOverallColor(reputation.overall_score || 0)}`}>
          {reputation.overall_score || 0}
        </span>
        <span className="text-xs text-gray-400">/ 100</span>
        {badges.length > 0 && (
          <div className="flex gap-1">
            {badges.slice(0, 2).map((badge, index) => (
              <span
                key={index}
                title={badge}
                className="inline-flex items-center rounded-full bg-purple-50 p-1.5"
              >
                {badgeIcons[badge] || <Award className="h-3.5 w-3.5 text-purple-500" />}
              </span>
            ))}
          </div>
        )}
      </div>
    );
  }

  return (
    <div className="rounded-2xl border border-gray-100 bg-white p-6 shadow-sm">
      <h3 className="mb-4 text-xs font-medium uppercase tracking-wider text-gray-400">
        Reputation Overview
      </h3>

      <div className="mb-6 flex items-end justify-between">
        <div>
          <p className="text-xs text-gray-400">Overall Score</p>
          <span className={`text-4xl font-bold ${getOverallColor(reputation.overall_score || 0)}`}>
            {reputation.overall_score || 0}
          </span>
          <span className="ml-1 text-sm text-gray-400">/ 100</span>
        </div>
        {averageRating !== undefined && (
          <div className="flex flex-col items-end gap-1">
            <p className="text-xs text-gray-400">Avg Rating</p>
            <StarRating rating={Math.round(averageRating)} />
          </div>
        )}
      </div>

      <div className="mb-6">
        <h4 className="mb-3 text-xs font-medium uppercase tracking-wider text-gray-400">Breakdown</h4>
        <ProgressBar label="Communication" score={reputation.communication_score || 0} />
        <ProgressBar label="Technical" score={reputation.technical_score || 0} />
        <ProgressBar label="Collaboration" score={reputation.collaboration_score || 0} />
      </div>

      <div>
        <h4 className="mb-3 text-xs font-medium uppercase tracking-wider text-gray-400">Badges</h4>
        {badges.length > 0 ? (
          <div className="flex flex-wrap gap-2">
            {badges.map((badge, index) => (
              <span
                key={index}
                className="inline-flex items-center gap-1.5 rounded-full bg-purple-50 px-3 py-1 text-xs font-medium text-purple-700"
              >
                {badgeIcons[badge] || <Award className="h-3 w-3" />}
                {badge}
              </span>
            ))}
          </div>
        ) : (
          <p className="text-sm text-gray-400">No badges earned yet.</p>
        )}
      </div>
    </div>
  );
};

export default ReputationDisplay;
