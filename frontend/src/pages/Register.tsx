import React, { useState, useEffect } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useForm } from 'react-hook-form';
import { useAuth } from '../contexts/AuthContext';
import toast from 'react-hot-toast';
import { FiUser, FiAtSign, FiLock, FiEye, FiEyeOff, FiLoader, FiGithub } from 'react-icons/fi';

interface RegisterFormData {
  full_name: string;
  username: string;
  email: string;
  password: string;
}

const Register: React.FC = () => {
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<RegisterFormData>();
  const { register: authRegister, loading, isAuthenticated, loginWithGoogle, loginWithGitHub } = useAuth();
  const navigate = useNavigate();

  const [showPassword, setShowPassword] = useState(false);

  useEffect(() => {
    if (isAuthenticated) {
      navigate('/dashboard', { replace: true });
    }
  }, [isAuthenticated, navigate]);

  const onSubmit = async (data: RegisterFormData) => {
    try {
      await authRegister(data);
      toast.success('Registration successful! Welcome.');
      navigate('/dashboard');
    } catch (error: any) {
      toast.error(error.message || 'Registration failed.');
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 p-4">
      <div className="w-full max-w-sm">
        <div className="text-center mb-8">
            <Link to="/" className="inline-flex items-center justify-center">
                <img src="/assets/logo-full.png" alt="SkillSync" className="h-10 w-auto" />
            </Link>
          <h1 className="mt-4 text-3xl font-bold text-text-primary">Create Your Account</h1>
          <p className="text-text-secondary mt-2">Join the community and start collaborating.</p>
        </div>

        <div className="bg-card-bg p-8 rounded-large shadow-card">
          <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-text-secondary mb-1" htmlFor="full_name">Full Name</label>
              <div className="relative">
                <FiUser className="pointer-events-none w-5 h-5 text-gray-400 absolute top-1/2 transform -translate-y-1/2 left-3" />
                <input id="full_name" type="text" {...register('full_name', { required: 'Full name is required' })}
                  className={`w-full pl-10 pr-3 py-2.5 border rounded-lg focus:ring-primary focus:border-primary ${errors.full_name ? 'border-red-500' : 'border-border'}`} />
              </div>
              {errors.full_name && <p className="text-red-500 text-xs mt-1">{errors.full_name.message}</p>}
            </div>

            <div>
                <label className="block text-sm font-medium text-text-secondary mb-1" htmlFor="username">Username</label>
                <div className="relative">
                    <FiAtSign className="pointer-events-none w-5 h-5 text-gray-400 absolute top-1/2 transform -translate-y-1/2 left-3" />
                    <input id="username" type="text" {...register('username', { required: 'Username is required' })}
                    className={`w-full pl-10 pr-3 py-2.5 border rounded-lg focus:ring-primary focus:border-primary ${errors.username ? 'border-red-500' : 'border-border'}`} />
                </div>
                {errors.username && <p className="text-red-500 text-xs mt-1">{errors.username.message}</p>}
            </div>

            <div>
                <label className="block text-sm font-medium text-text-secondary mb-1" htmlFor="email">Email Address</label>
                <div className="relative">
                    <FiAtSign className="pointer-events-none w-5 h-5 text-gray-400 absolute top-1/2 transform -translate-y-1/2 left-3" />
                    <input id="email" type="email" {...register('email', { required: 'Email is required', pattern: /^\S+@\S+$/i })}
                    className={`w-full pl-10 pr-3 py-2.5 border rounded-lg focus:ring-primary focus:border-primary ${errors.email ? 'border-red-500' : 'border-border'}`} />
                </div>
                {errors.email && <p className="text-red-500 text-xs mt-1">Please enter a valid email.</p>}
            </div>

            <div>
                <label className="block text-sm font-medium text-text-secondary mb-1" htmlFor="password">Password</label>
                <div className="relative">
                    <FiLock className="pointer-events-none w-5 h-5 text-gray-400 absolute top-1/2 transform -translate-y-1/2 left-3" />
                    <input id="password" type={showPassword ? 'text' : 'password'} {...register('password', { required: 'Password is required', minLength: { value: 6, message: 'Password must be at least 6 characters.' } })}
                    className={`w-full pl-10 pr-10 py-2.5 border rounded-lg focus:ring-primary focus:border-primary ${errors.password ? 'border-red-500' : 'border-border'}`} />
                    <button type="button" onClick={() => setShowPassword(!showPassword)}
                    className="absolute top-1/2 transform -translate-y-1/2 right-3 text-gray-400 hover:text-gray-600">
                    {showPassword ? <FiEyeOff /> : <FiEye />}
                    </button>
                </div>
                {errors.password && <p className="text-red-500 text-xs mt-1">{errors.password.message}</p>}
            </div>
            
            <button type="submit" disabled={loading} className="w-full btn btn-primary py-3 !mt-6">
              {loading ? <FiLoader className="h-5 w-5 animate-spin mx-auto" /> : 'Create Account'}
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
          Already have an account?{' '}
          <Link to="/login" className="font-semibold text-primary hover:underline">
            Sign in
          </Link>
        </p>
      </div>
    </div>
  );
};

export default Register;
