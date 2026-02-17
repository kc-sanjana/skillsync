import React, { useState, useEffect } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import toast from 'react-hot-toast';
import { FiMail, FiLock, FiEye, FiEyeOff, FiLoader, FiGithub } from 'react-icons/fi';

const Login: React.FC = () => {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);

  const { login, loading, isAuthenticated, loginWithGoogle, loginWithGitHub } = useAuth();
  const navigate = useNavigate();

  useEffect(() => {
    if (isAuthenticated) {
      navigate('/dashboard', { replace: true });
    }
  }, [isAuthenticated, navigate]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!email || !password) {
      toast.error('Please enter both email and password.');
      return;
    }
    try {
      await login({ email, password });
      toast.success('Logged in successfully!');
    } catch (error: any) {
      toast.error(error.message || 'Login failed. Please check your credentials.');
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 p-4">
      <div className="w-full max-w-sm">
        <div className="text-center mb-8">
            <Link to="/" className="inline-flex items-center justify-center">
                <img src="/assets/logo-full.png" alt="SkillSync" className="h-10 w-auto" />
            </Link>
          <h1 className="mt-4 text-3xl font-bold text-text-primary">Welcome Back</h1>
          <p className="text-text-secondary mt-2">Sign in to continue your skill-sharing journey.</p>
        </div>

        <div className="bg-card-bg p-8 rounded-large shadow-card">
          <form onSubmit={handleSubmit} className="space-y-6">
            <div>
              <label htmlFor="email" className="block text-sm font-medium text-text-secondary mb-1">
                Email Address
              </label>
              <div className="relative">
                <FiMail className="pointer-events-none w-5 h-5 text-gray-400 absolute top-1/2 transform -translate-y-1/2 left-3" />
                <input
                  id="email"
                  type="email"
                  required
                  autoComplete="email"
                  placeholder="you@example.com"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  className="w-full pl-10 pr-3 py-2.5 border border-border rounded-lg focus:ring-primary focus:border-primary"
                />
              </div>
            </div>

            <div>
              <label htmlFor="password"  className="block text-sm font-medium text-text-secondary mb-1">
                Password
              </label>
              <div className="relative">
                 <FiLock className="pointer-events-none w-5 h-5 text-gray-400 absolute top-1/2 transform -translate-y-1/2 left-3" />
                <input
                  id="password"
                  type={showPassword ? 'text' : 'password'}
                  required
                  autoComplete="current-password"
                  placeholder="••••••••"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                   className="w-full pl-10 pr-10 py-2.5 border border-border rounded-lg focus:ring-primary focus:border-primary"
                />
                <button
                  type="button"
                  onClick={() => setShowPassword(!showPassword)}
                  className="absolute top-1/2 transform -translate-y-1/2 right-3 text-gray-400 hover:text-gray-600"
                >
                  {showPassword ? <FiEyeOff /> : <FiEye />}
                </button>
              </div>
            </div>

            <button type="submit" disabled={loading} className="w-full btn btn-primary py-3">
              {loading ? <FiLoader className="h-5 w-5 animate-spin mx-auto" /> : 'Sign In'}
            </button>
          </form>

          <div className="my-6 flex items-center gap-3">
            <div className="h-px flex-1 bg-border" />
            <span className="text-xs text-text-secondary">OR</span>
            <div className="h-px flex-1 bg-border" />
          </div>

          <div className="space-y-3">
             <button onClick={loginWithGoogle} className="w-full flex items-center justify-center gap-3 py-2.5 border border-border rounded-lg hover:bg-gray-50 transition-colors">
                <img src="https://www.svgrepo.com/show/475656/google-color.svg" alt="Google" className="w-5 h-5" />
                <span className="text-sm font-medium text-text-secondary">Continue with Google</span>
            </button>
             <button onClick={loginWithGitHub} className="w-full flex items-center justify-center gap-3 py-2.5 border border-border rounded-lg hover:bg-gray-50 transition-colors">
                <FiGithub className="w-5 h-5 text-text-secondary" />
                <span className="text-sm font-medium text-text-secondary">Continue with GitHub</span>
            </button>
          </div>
        </div>

         <p className="mt-8 text-center text-sm text-text-secondary">
            Don't have an account?{' '}
            <Link to="/register" className="font-semibold text-primary hover:underline">
              Sign up
            </Link>
          </p>
      </div>
    </div>
  );
};

export default Login;
