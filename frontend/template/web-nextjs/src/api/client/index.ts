import {
  type AxiosError,
  type AxiosInstance,
  type AxiosResponse,
  type InternalAxiosRequestConfig,
} from "axios";
import { createApiClient, type ApiErrorResponse } from "@devkit/shared";

// ---------------------------------------------------------------------------
// Web-specific Axios instance with localStorage auth token
// ---------------------------------------------------------------------------
const API_BASE_URL =
  process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:5555";

const apiClient: AxiosInstance = createApiClient(API_BASE_URL);

// Override request interceptor to attach localStorage token
apiClient.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    const token =
      typeof window !== "undefined"
        ? localStorage.getItem("auth_token")
        : null;
    if (token && config.headers) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error: AxiosError) => Promise.reject(error),
);

// Override response interceptor for 401 handling
apiClient.interceptors.response.use(
  (response: AxiosResponse) => response,
  (error: AxiosError<ApiErrorResponse>) => {
    if (error.response?.status === 401 && typeof window !== "undefined") {
      localStorage.removeItem("auth_token");
    }
    return Promise.reject(error);
  },
);

export { apiClient, type ApiErrorResponse };
export default apiClient;
