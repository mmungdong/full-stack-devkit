import { type AxiosInstance, type AxiosRequestConfig } from "axios";
import { type ZodSchema } from "zod";
import { type ApiErrorResponse } from "./client";

// ---------------------------------------------------------------------------
// Generic typed API request with Zod response validation
// ---------------------------------------------------------------------------
export interface ApiRequestResult<T> {
  data: T | null;
  error: Error | null;
}

/**
 * Perform a validated API request using any Axios instance.
 *
 * 1. Send request via the provided client
 * 2. Validate the response against the provided Zod schema
 * 3. Return a safe [data, error] tuple
 */
export async function apiRequest<T>(
  client: AxiosInstance,
  method: "get" | "post" | "put" | "patch" | "delete",
  url: string,
  schema: ZodSchema<T>,
  config?: AxiosRequestConfig & { data?: unknown },
): Promise<ApiRequestResult<T>> {
  try {
    const response = await client.request<T>({ method, url, ...config });
    const result = schema.safeParse(response.data);

    if (!result.success) {
      const errors = result.error.issues.map(
        (issue) => `${issue.path.join(".")}: ${issue.message}`,
      );
      return {
        data: null,
        error: new Error(`Validation failed:\n${errors.join("\n")}`),
      };
    }

    return { data: result.data, error: null };
  } catch (err) {
    const error =
      err instanceof Error ? err : new Error("An unknown error occurred");

    // Enrich error message from backend response
    if (typeof err === "object" && err !== null && "response" in err) {
      const axiosErr = err as { response?: { data?: ApiErrorResponse } };
      if (axiosErr.response?.data?.message) {
        return {
          data: null,
          error: new Error(axiosErr.response.data.message),
        };
      }
    }

    return { data: null, error };
  }
}

// ---------------------------------------------------------------------------
// Convenience helpers — accept client as first arg
// ---------------------------------------------------------------------------
export function apiGet<T>(
  client: AxiosInstance,
  url: string,
  schema: ZodSchema<T>,
  config?: AxiosRequestConfig,
) {
  return apiRequest<T>(client, "get", url, schema, config);
}

export function apiPost<T>(
  client: AxiosInstance,
  url: string,
  schema: ZodSchema<T>,
  data?: unknown,
  config?: AxiosRequestConfig,
) {
  return apiRequest<T>(client, "post", url, schema, { ...config, data });
}

export function apiPut<T>(
  client: AxiosInstance,
  url: string,
  schema: ZodSchema<T>,
  data?: unknown,
  config?: AxiosRequestConfig,
) {
  return apiRequest<T>(client, "put", url, schema, { ...config, data });
}

export function apiPatch<T>(
  client: AxiosInstance,
  url: string,
  schema: ZodSchema<T>,
  data?: unknown,
  config?: AxiosRequestConfig,
) {
  return apiRequest<T>(client, "patch", url, schema, { ...config, data });
}

export function apiDelete<T>(
  client: AxiosInstance,
  url: string,
  schema: ZodSchema<T>,
  config?: AxiosRequestConfig,
) {
  return apiRequest<T>(client, "delete", url, schema, config);
}
