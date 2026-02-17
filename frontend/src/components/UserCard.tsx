import React from 'react';
import { Link } from 'react-router-dom';
import { User } from '../types';
import { FiStar, FiArrowRight } from 'react-icons/fi';

interface UserCardProps {
  user: User;
}

const UserCard: React.FC<UserCardProps> = ({ user }) => {
  const avatarUrl = `https://ui-avatars.com/api/?name=${encodeURIComponent(
    user.full_name || user.username
  )}&background=random&color=fff&bold=true&rounded=true`;

  const userSkills = [...(user.skills_teach || []), ...(user.skills_learn || [])].slice(0, 3);

  return (
    <div className="bg-card-bg rounded-large shadow-card p-6 flex flex-col h-full transition-transform duration-300 ease-in-out hover:-translate-y-1 hover:shadow-card-hover">
      <div className="flex items-center gap-4">
        <img src={avatarUrl} alt={user.username} className="h-14 w-14 rounded-full" />
        <div className="flex-1 overflow-hidden">
          <Link to={`/profile/${user.id}`} className="font-semibold text-text-primary hover:text-primary transition-colors">
            <span className="block truncate">{user.full_name || user.username}</span>
          </Link>
          {user.reputation_score > 0 && (
            <div className="mt-1 flex items-center gap-1 text-xs text-text-secondary">
              <FiStar className="h-3.5 w-3.5 text-yellow-500" />
              <span className="font-medium">{user.reputation_score.toFixed(1)} Reputation</span>
            </div>
          )}
        </div>
      </div>

      {userSkills.length > 0 && (
        <div className="mt-4 flex flex-wrap gap-2">
          {userSkills.map((skill, index) => (
            <span key={index} className="bg-primary/10 text-primary text-xs font-medium px-3 py-1 rounded-full">
              {skill}
            </span>
          ))}
        </div>
      )}

      <div className="mt-auto pt-6">
        <Link
          to={`/profile/${user.id}`}
          className="group w-full bg-gradient-to-r from-gradient-from to-gradient-to text-white font-semibold px-4 py-2.5 rounded-lg flex items-center justify-center gap-2 transition-all duration-300 ease-in-out hover:shadow-lg hover:shadow-primary/30"
        >
          <span>Connect</span>
          <FiArrowRight className="h-4 w-4 transition-transform duration-300 group-hover:translate-x-1" />
        </Link>
      </div>
    </div>
  );
};

export default UserCard;
