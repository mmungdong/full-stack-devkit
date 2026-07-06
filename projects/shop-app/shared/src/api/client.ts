import axios, {
  type AxiosError,
  type AxiosInstance,
  type AxiosResponse,
  type InternalAxiosRequestConfig,
} from "axios";

// ---------------------------------------------------------------------------
// API error shape returned by the Go backend (osbuilder / onexstack)
// ---------------------------------------------------------------------------
export interface ApiErrorResponse {
  code: number;
  message: string;
  details?: string;
}

// ---------------------------------------------------------------------------
// Create a pre-configured Axios instance
// ---------------------------------------------------------------------------
function createApiClient(baseURL: string): AxiosInstance {
  const instance = axios.create({
    baseURL,
    timeout: 15_000,
    headers: {
      "Content-Type": "application/json",
    },
  });

  // ---- Request interceptor ----
  instance.interceptors.request.use(
    (config: InternalAxiosRequestConfig) => {
      // Note: Token injection is platform-specific.
      // Web → localStorage, RN → SecureStorage (override in each app)
      return config;
    },
    (error: AxiosError) => Promise.reject(error),
  );

  // ---- Response interceptor ----
  instance.interceptors.response.use(
    (response: AxiosResponse) => response,
    (error: AxiosError<ApiErrorResponse>) => {
      if (error.response) {
        const { status } = error.response;

        // Handle 401 — clear auth state
        if (status === 401) {
          // Note: Auth state cleanup is platform-specific.
          // Override in each app's interceptor.
        }
      }

      return Promise.reject(error);
    },
  );

  return instance;
}

export { createApiClient };
