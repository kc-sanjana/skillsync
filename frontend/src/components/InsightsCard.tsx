import React from 'react';
import { Sparkles, Lightbulb, BookOpen, Star } from 'lucide-react';
import { PairingInsights } from '../types';

interface InsightsCardProps {
  insights: PairingInsights;
}

const InsightsCard: React.FC<InsightsCardProps> = ({ insights }) => {
  return (
    <div className="rounded-2xl border border-purple-200 bg-gradient-to-br from-purple-50 via-pink-50 to-indigo-50 p-6">
      <h3 className="mb-4 flex items-center gap-2 text-lg font-bold text-gray-900">
        <Sparkles className="h-5 w-5 text-indigo-400" />
        Why You Match
      </h3>
      <p className="text-sm leading-relaxed text-gray-500">{insights.overall_reasoning}</p>

      <div className="mt-5 grid grid-cols-1 gap-5 md:grid-cols-2">
        <div>
          <h4 className="mb-2 flex items-center gap-2 text-sm font-medium text-gray-600">
            <Lightbulb className="h-4 w-4 text-amber-400" />
            Collaboration Ideas
          </h4>
          {insights.collaboration_ideas && insights.collaboration_ideas.length > 0 ? (
            <ul className="space-y-1.5">
              {insights.collaboration_ideas.map((item, index) => (
                <li key={index} className="flex items-start gap-2 text-sm text-gray-500">
                  <span className="mt-1.5 h-1 w-1 flex-shrink-0 rounded-full bg-purple-400" />
                  {item}
                </li>
              ))}
            </ul>
          ) : (
            <p className="text-sm text-gray-400">No specific collaboration ideas generated.</p>
          )}
        </div>

        <div>
          <h4 className="mb-2 flex items-center gap-2 text-sm font-medium text-gray-600">
            <BookOpen className="h-4 w-4 text-emerald-400" />
            Learning Opportunities
          </h4>
          {insights.learning_opportunities && insights.learning_opportunities.length > 0 ? (
            <ul className="space-y-1.5">
              {insights.learning_opportunities.map((item, index) => (
                <li key={index} className="flex items-start gap-2 text-sm text-gray-500">
                  <span className="mt-1.5 h-1 w-1 flex-shrink-0 rounded-full bg-emerald-400" />
                  {item}
                </li>
              ))}
            </ul>
          ) : (
            <p className="text-sm text-gray-400">No specific learning opportunities identified.</p>
          )}
        </div>
      </div>

      {insights.recommendation && (
        <div className="mt-5 rounded-xl bg-white/70 px-4 py-3">
          <p className="flex items-center gap-2 text-sm text-gray-500">
            <Star className="h-4 w-4 flex-shrink-0 text-amber-400" />
            <span>
              <span className="font-medium text-gray-600">Recommendation:</span>{' '}
              {insights.recommendation}
            </span>
          </p>
        </div>
      )}
    </div>
  );
};

export default InsightsCard;
