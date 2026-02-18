// src/services/auth.ts
import axios, { AxiosError } from "axios";
import { User } from "../types";

const API_BASE_URL =
  import.meta.env.VITE_API_BASE_URL || "http://localhost:8080/api/v1";

interface AuthResponse {
  token: string;
  user: User;
}

const authService = {
  // 📝 REGISTER
  async register(data: any): Promise<AuthResponse> {
    try {
      const response = await axios.post<AuthResponse>(
        `${API_BASE_URL}/auth/register`,
        data
      );

      const { token, user } = response.data;

      // ✅ Save token
      localStorage.setItem("jwt_token", token);

      return { token, user };
    } catch (error) {
      const err = error as AxiosError<any>;
      throw new Error(
        err.response?.data?.message ||
          err.response?.data?.error ||
          "Registration failed"
      );
    }
  },

  // 🔐 LOGIN
  async login(data: any): Promise<AuthResponse> {
    try {
      const response = await axios.post<AuthResponse>(
        `${API_BASE_URL}/auth/login`,
        data
      );

      const { token, user } = response.data;

      // ✅ Save token
      localStorage.setItem("jwt_token", token);

      return { token, user };
    } catch (error) {
      const err = error as AxiosError<any>;
      throw new Error(
        err.response?.data?.message ||
          err.response?.data?.error ||
          "Invalid email or password"
      );
    }
  },

  // 🚪 LOGOUT
  async logout(): Promise<void> {
    localStorage.removeItem("jwt_token");
  },

  // 👤 GET CURRENT USER  ✅ FIXED ENDPOINT
  async getCurrentUser(): Promise<User> {
    const token = localStorage.getItem("jwt_token");

    if (!token) {
      throw new Error("No token found");
    }

    try {
      const response = await axios.get<User>(
        `${API_BASE_URL}/users/me`,
        {
          headers: {
            Authorization: `Bearer ${token}`,
          },
        }
      );

      return response.data;
    } catch (error) {
      const err = error as AxiosError<any>;

      // ❗ Token invalid → clear it
      if (err.response?.status === 401 || err.response?.status === 404) {
        localStorage.removeItem("jwt_token");
      }

      throw new Error(
        err.response?.data?.message ||
          err.response?.data?.error ||
          "Failed to fetch user"
      );
    }
  },
};

export default authService;

