import React, {
  createContext,
  useState,
  useEffect,
  useContext,
  ReactNode,
  useCallback,
} from "react";
import { User } from "../types";
import authService from "../services/auth";
import toast from "react-hot-toast";

interface AuthContextType {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  loading: boolean;
  login: (credentials: any) => Promise<void>;
  register: (userData: any) => Promise<void>;
  logout: () => void;
  refreshUser: () => Promise<void>;
  loginWithGoogle: () => void;
  loginWithGitHub: () => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

interface AuthProviderProps {
  children: ReactNode;
}

export const AuthProvider: React.FC<AuthProviderProps> = ({ children }) => {
  // 🔐 Load token from storage OR from OAuth redirect (synchronous to avoid race)
  const [token, setToken] = useState<string | null>(() => {
    const params = new URLSearchParams(window.location.search);
    const oauthToken = params.get("token");
    if (oauthToken) {
      localStorage.setItem("jwt_token", oauthToken);
      window.history.replaceState({}, "", window.location.pathname);
      return oauthToken;
    }
    return localStorage.getItem("jwt_token");
  });

  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  // ✅ Authenticated if token exists AND user loaded
  const isAuthenticated = !!token && !!user;

  // 🌐 Handle OAuth error redirects
  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    const error = params.get("error");
    if (error) {
      const messages: Record<string, string> = {
        invalid_state: "OAuth session expired. Please try again.",
        no_code: "Authentication was cancelled.",
        oauth_failed: "OAuth login failed. Please try again.",
        token_failed: "Something went wrong. Please try again.",
      };
      toast.error(messages[error] || "Login failed.");
      window.history.replaceState({}, "", window.location.pathname);
    }
  }, []);

  // 🔄 Fetch current user
  const refreshUser = useCallback(async () => {
    if (!token) {
      setUser(null);
      setLoading(false);
      return;
    }

    try {
      const currentUser = await authService.getCurrentUser();
      setUser(currentUser);
    } catch (err) {
      console.warn("⚠️ Token invalid → logging out");
      logout();
    } finally {
      setLoading(false);
    }
  }, [token]);

  useEffect(() => {
    refreshUser();
  }, [refreshUser]);

  // 🔐 LOGIN
  const login = async (credentials: any) => {
    setLoading(true);
    try {
      const { token: newToken, user: newUser } =
        await authService.login(credentials);

      // ✅ Save token first
      localStorage.setItem("jwt_token", newToken);
      setToken(newToken);

      // ✅ Set user immediately
      setUser(newUser);

    } finally {
      setLoading(false);
    }
  };

  // 📝 REGISTER
  const register = async (userData: any) => {
    setLoading(true);
    try {
      const { token: newToken, user: newUser } =
        await authService.register(userData);

      localStorage.setItem("jwt_token", newToken);
      setToken(newToken);
      setUser(newUser);
    } finally {
      setLoading(false);
    }
  };

  // 🚪 LOGOUT
  const logout = useCallback(() => {
    localStorage.removeItem("jwt_token");
    setToken(null);
    setUser(null);
  }, []);

  // 🌐 SOCIAL LOGINS
  const API_BASE = import.meta.env.VITE_API_BASE_URL || "http://localhost:8080/api/v1";

  const loginWithGoogle = () => {
    window.location.href = `${API_BASE}/auth/google/login`;
  };

  const loginWithGitHub = () => {
    window.location.href = `${API_BASE}/auth/github/login`;
  };

  const value: AuthContextType = {
    user,
    token,
    isAuthenticated,
    loading,
    login,
    register,
    logout,
    refreshUser,
    loginWithGoogle,
    loginWithGitHub,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
};

// 🔌 Hook
export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error("useAuth must be used within AuthProvider");
  }
  return context;
};
