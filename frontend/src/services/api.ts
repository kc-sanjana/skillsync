// src/services/api.ts
import axios, { AxiosInstance, AxiosError, AxiosResponse } from "axios";
import { APIResponse } from "../types";

/**
 * ✅ Vite-compatible environment variable
 * Uses VITE_API_BASE_URL from .env
 */
const API_BASE_URL =
  import.meta.env.VITE_API_BASE_URL || "http://localhost:8080/api/v1";

/**
 * Axios instance
 */
const apiClient: AxiosInstance = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    "Content-Type": "application/json",
  },
});

/**
 * 🔐 Request interceptor: attach JWT
 */
apiClient.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem("jwt_token");
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => Promise.reject(error)
);

/**
 * ⚠️ Response interceptor: handle auth errors
 */
apiClient.interceptors.response.use(
  (response) => response,
  async (error: AxiosError) => {
    if (error.response?.status === 401) {
      console.warn("Unauthorized — removing token");
      localStorage.removeItem("jwt_token");
      // optional redirect:
      // window.location.href = "/login";
    }
    return Promise.reject(error);
  }
);

/**
 * 📦 Typed helper methods
 */
const api = {
  get: async <T>(url: string, config?: object): Promise<APIResponse<T>> => {
    try {
      const response: AxiosResponse<APIResponse<T>> =
        await apiClient.get(url, config);
      return response.data;
    } catch (error) {
      const err = error as AxiosError<APIResponse<T>>;
      return (
        err.response?.data || {
          success: false,
          error: { code: "network_error", message: err.message },
        }
      );
    }
  },

  post: async <T>(url: string, data?: object): Promise<APIResponse<T>> => {
    try {
      const response: AxiosResponse<APIResponse<T>> =
        await apiClient.post(url, data);
      return response.data;
    } catch (error) {
      const err = error as AxiosError<APIResponse<T>>;
      return (
        err.response?.data || {
          success: false,
          error: { code: "network_error", message: err.message },
        }
      );
    }
  },

  put: async <T>(url: string, data?: object): Promise<APIResponse<T>> => {
    try {
      const response: AxiosResponse<APIResponse<T>> =
        await apiClient.put(url, data);
      return response.data;
    } catch (error) {
      const err = error as AxiosError<APIResponse<T>>;
      return (
        err.response?.data || {
          success: false,
          error: { code: "network_error", message: err.message },
        }
      );
    }
  },

  delete: async <T>(url: string): Promise<APIResponse<T>> => {
    try {
      const response: AxiosResponse<APIResponse<T>> =
        await apiClient.delete(url);
      return response.data;
    } catch (error) {
      const err = error as AxiosError<APIResponse<T>>;
      return (
        err.response?.data || {
          success: false,
          error: { code: "network_error", message: err.message },
        }
      );
    }
  },
};

export default api;

