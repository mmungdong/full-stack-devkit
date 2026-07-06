import { type AxiosRequestConfig } from "axios";
import { type ZodSchema } from "zod";
import {
  apiGet as sharedGet,
  apiPost as sharedPost,
  apiPut as sharedPut,
  apiPatch as sharedPatch,
  apiDelete as sharedDelete,
} from "@devkit/shared";
import { apiClient } from "@/api/client";

// ---------------------------------------------------------------------------
// Re-export shared request helpers, pre-bound to the web API client
// ---------------------------------------------------------------------------

export function webGet<T>(
  url: string,
  schema: ZodSchema<T>,
  config?: AxiosRequestConfig,
) {
  return sharedGet(apiClient, url, schema, config);
}

export function webPost<T>(
  url: string,
  schema: ZodSchema<T>,
  data?: unknown,
  config?: AxiosRequestConfig,
) {
  return sharedPost(apiClient, url, schema, data, config);
}

export function webPut<T>(
  url: string,
  schema: ZodSchema<T>,
  data?: unknown,
  config?: AxiosRequestConfig,
) {
  return sharedPut(apiClient, url, schema, data, config);
}

export function webPatch<T>(
  url: string,
  schema: ZodSchema<T>,
  data?: unknown,
  config?: AxiosRequestConfig,
) {
  return sharedPatch(apiClient, url, schema, data, config);
}

export function webDelete<T>(
  url: string,
  schema: ZodSchema<T>,
  config?: AxiosRequestConfig,
) {
  return sharedDelete(apiClient, url, schema, config);
}
