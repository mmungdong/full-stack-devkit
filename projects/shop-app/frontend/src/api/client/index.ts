import {
  type AxiosError,
  type AxiosInstance,
  type AxiosResponse,
  type InternalAxiosRequestConfig,
} from "axios";
import { createApiClient, type ApiErrorResponse } from "@devkit/shared";

// ---------------------------------------------------------------------------
// Web-specific Axios instance with zustand-persisted auth token
// ---------------------------------------------------------------------------
const API_BASE_URL =
  process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:5555/api";

const apiClient: AxiosInstance = createApiClient(API_BASE_URL);

// 请求拦截器：从 zustand persist 的 localStorage state 读 token
// （直接读 localStorage 而非导入 store，避免循环依赖：store 导入 apiClient，apiClient 不能反向导入 store）
apiClient.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    if (typeof window !== "undefined") {
      const raw = localStorage.getItem("shop-app-auth");
      if (raw) {
        try {
          const parsed = JSON.parse(raw);
          const token = parsed?.state?.token;
          if (token && config.headers) {
            config.headers.Authorization = `Bearer ${token}`;
          }
        } catch {
          // 解析失败忽略
        }
      }
    }
    return config;
  },
  (error: AxiosError) => Promise.reject(error),
);

// 响应拦截器：401 清理持久化 state 并跳登录页
apiClient.interceptors.response.use(
  (response: AxiosResponse) => response,
  (error: AxiosError<ApiErrorResponse>) => {
    if (error.response?.status === 401 && typeof window !== "undefined") {
      localStorage.removeItem("shop-app-auth");
      if (window.location.pathname !== "/login") {
        window.location.href = "/login";
      }
    }
    return Promise.reject(error);
  },
);

export { apiClient, type ApiErrorResponse };
export default apiClient;
