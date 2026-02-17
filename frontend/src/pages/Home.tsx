import React from 'react';
import { Link } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import { FiZap, FiMessageCircle, FiTrendingUp, FiArrowRight } from 'react-icons/fi';

const features = [
  {
    icon: FiZap,
    title: 'AI-Powered Matching',
    description: 'Our AI analyzes your skills and goals to find your perfect learning partner.',
  },
  {
    icon: FiMessageCircle,
    title: 'Real-Time Chat',
    description: 'Collaborate instantly with built-in messaging and live coding sessions.',
  },
  {
    icon: FiTrendingUp,
    title: 'Reputation System',
    description: 'Build credibility through peer ratings and climb the community leaderboard.',
  },
];

const Home: React.FC = () => {
  const { isAuthenticated } = useAuth();

  return (
    <div className="bg-gray-50 text-text-primary">
      {/* Header */}
      <header className="py-4 px-4 sm:px-6 lg:px-8">
        <nav className="flex items-center justify-between max-w-7xl mx-auto">
            <Link to="/" className="flex items-center">
                <img src="/assets/logo-full.png" alt="SkillSync" className="h-10 sm:h-12 w-auto" />
            </Link>
          <div>
            <Link to="/login" className="text-sm font-semibold text-text-secondary hover:text-primary transition-colors mr-6">
              Log in
            </Link>
            <Link to={isAuthenticated ? '/dashboard' : '/register'} className="btn btn-primary">
              {isAuthenticated ? 'Dashboard' : 'Get Started'}
            </Link>
          </div>
        </nav>
      </header>

      <main>
        {/* Hero Section */}
        <section className="text-center py-20 sm:py-28 px-4">
            <h1 className="text-4xl sm:text-5xl lg:text-6xl font-extrabold text-text-primary tracking-tight">
              Find Your Perfect <br/>
              <span className="bg-gradient-to-r from-gradient-from to-gradient-to bg-clip-text text-transparent">Learning Partner</span>
            </h1>
            <p className="max-w-2xl mx-auto mt-6 text-lg text-text-secondary">
              SkillSync uses AI to match you with the ideal partner for skill exchange.
              Learn, teach, and grow together in a community of lifelong learners.
            </p>
            <div className="mt-8 flex justify-center gap-4">
              <Link to={isAuthenticated ? '/dashboard' : '/register'} className="group btn btn-primary px-8 py-3 text-base flex items-center gap-2">
                <span>{isAuthenticated ? 'Go to Dashboard' : 'Get Started For Free'}</span>
                <FiArrowRight className="h-5 w-5 transition-transform duration-300 group-hover:translate-x-1" />
              </Link>
            </div>
        </section>

        {/* Features Section */}
        <section className="py-20 sm:py-28 bg-white">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
            <div className="text-center">
              <h2 className="text-3xl font-bold text-text-primary sm:text-4xl">
                Everything you need to level up
              </h2>
              <p className="mt-4 max-w-2xl mx-auto text-lg text-text-secondary">
                Powerful features designed to help you learn faster and build real connections.
              </p>
            </div>
            <div className="mt-16 grid grid-cols-1 gap-10 sm:grid-cols-2 lg:grid-cols-3">
              {features.map(({ icon: Icon, title, description }) => (
                <div key={title} className="text-center p-8 bg-card-bg rounded-large shadow-card">
                  <div className="flex items-center justify-center h-12 w-12 rounded-lg bg-primary/10 mx-auto">
                    <Icon className="h-6 w-6 text-primary" />
                  </div>
                  <h3 className="mt-6 text-xl font-bold text-text-primary">{title}</h3>
                  <p className="mt-2 text-text-secondary">{description}</p>
                </div>
              ))}
            </div>
          </div>
        </section>
      </main>

      {/* Footer */}
      <footer className="py-8 bg-white border-t border-border">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 text-center text-sm text-text-secondary">
            &copy; {new Date().getFullYear()} SkillSync. All rights reserved.
        </div>
      </footer>
    </div>
  );
};

export default Home;
