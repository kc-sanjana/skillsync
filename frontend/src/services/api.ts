// src/services/api.ts
import axios, { AxiosInstance, AxiosError, AxiosResponse } from "axios";
import { APIResponse } from "../types";

const API_BASE_URL =
  import.meta.env.VITE_API_BASE_URL || "http://localhost:8080/api";

const apiClient: AxiosInstance = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    "Content-Type": "application/json",
  },
});

// Request interceptor: attach JWT
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

// Response interceptor: handle auth errors
apiClient.interceptors.response.use(
  (response) => response,
  async (error: AxiosError) => {
    if (error.response?.status === 401) {
      console.warn("Unauthorized — removing token");
      localStorage.removeItem("jwt_token");
    }
    return Promise.reject(error);
  }
);

/**
 * Normalize backend responses into APIResponse format.
 * The backend returns raw data (not wrapped in { success, data }),
 * so we wrap successful responses ourselves.
 */
function wrapResponse<T>(data: any): APIResponse<T> {
  // If the backend already returns { success, data } format, pass through
  if (data && typeof data.success === "boolean") {
    return data as APIResponse<T>;
  }
  // Otherwise wrap the raw response
  return { success: true, data: data as T };
}

function wrapError<T>(error: unknown): APIResponse<T> {
  const err = error as AxiosError<any>;
  const responseData = err.response?.data;

  // If backend returned an error object with { error: "..." }
  if (responseData?.error && typeof responseData.error === "string") {
    return {
      success: false,
      error: { code: "api_error", message: responseData.error },
    };
  }

  return {
    success: false,
    error: { code: "network_error", message: err.message },
  };
}

const api = {
  get: async <T>(url: string, config?: object): Promise<APIResponse<T>> => {
    try {
      const response: AxiosResponse = await apiClient.get(url, config);
      return wrapResponse<T>(response.data);
    } catch (error) {
      return wrapError<T>(error);
    }
  },

  post: async <T>(url: string, data?: object): Promise<APIResponse<T>> => {
    try {
      const response: AxiosResponse = await apiClient.post(url, data);
      return wrapResponse<T>(response.data);
    } catch (error) {
      return wrapError<T>(error);
    }
  },

  put: async <T>(url: string, data?: object): Promise<APIResponse<T>> => {
    try {
      const response: AxiosResponse = await apiClient.put(url, data);
      return wrapResponse<T>(response.data);
    } catch (error) {
      return wrapError<T>(error);
    }
  },

  delete: async <T>(url: string): Promise<APIResponse<T>> => {
    try {
      const response: AxiosResponse = await apiClient.delete(url);
      return wrapResponse<T>(response.data);
    } catch (error) {
      return wrapError<T>(error);
    }
  },
};

export default api;
