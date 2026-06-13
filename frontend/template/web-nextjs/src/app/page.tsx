"use client";

import { useCallback, useEffect, useState } from "react";
import { webGet } from "@/api/request";
import { API, APP_NAME } from "@/constants";
import { HealthzResponseSchema, type HealthzResponse } from "@/schemas";

export default function Home() {
  const [health, setHealth] = useState<HealthzResponse | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  const checkHealth = useCallback(async () => {
    setLoading(true);
    setError(null);
    const result = await webGet(API.HEALTHZ, HealthzResponseSchema);
    if (result.data) {
      setHealth(result.data);
    } else if (result.error) {
      setError(result.error.message);
    }
    setLoading(false);
  }, []);

  // Fetch health on mount using a callback ref pattern
  // to avoid the "setState in effect" lint rule
  useEffect(() => {
    let cancelled = false;

    async function fetchHealth() {
      setLoading(true);
      setError(null);
      const result = await webGet(API.HEALTHZ, HealthzResponseSchema);
      if (cancelled) return;
      if (result.data) {
        setHealth(result.data);
      } else if (result.error) {
        setError(result.error.message);
      }
      setLoading(false);
    }

    fetchHealth();

    return () => {
      cancelled = true;
    };
  }, []);

  return (
    <div className="flex flex-col flex-1 items-center justify-center bg-zinc-50 font-sans dark:bg-black">
      <main className="flex flex-1 w-full max-w-3xl flex-col items-center justify-between py-32 px-16 bg-white dark:bg-black sm:items-start">
        <div className="flex flex-col items-center gap-6 text-center sm:items-start sm:text-left">
          <h1 className="max-w-md text-3xl font-semibold leading-10 tracking-tight text-black dark:text-zinc-50">
            {APP_NAME}
          </h1>
          <p className="max-w-lg text-lg leading-8 text-zinc-600 dark:text-zinc-400">
            Full-stack application powered by Go (osbuilder) backend and Next.js
            static frontend.
          </p>
        </div>

        {/* Health Check Section */}
        <div className="mt-8 w-full max-w-lg rounded-xl border border-zinc-200 p-6 dark:border-zinc-800">
          <h2 className="mb-4 text-xl font-semibold text-black dark:text-zinc-50">
            Backend Health Check
          </h2>

          <button
            onClick={checkHealth}
            disabled={loading}
            className="mb-4 rounded-lg bg-black px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-zinc-800 disabled:opacity-50 dark:bg-white dark:text-black dark:hover:bg-zinc-200"
          >
            {loading ? "Checking..." : "Check Health"}
          </button>

          {health && (
            <div className="rounded-lg bg-green-50 p-4 dark:bg-green-900/20">
              <p className="text-sm font-medium text-green-800 dark:text-green-400">
                ✅ Backend is healthy
              </p>
              <p className="mt-1 text-sm text-green-600 dark:text-green-500">
                Timestamp: {health.timestamp}
              </p>
            </div>
          )}

          {error && (
            <div className="rounded-lg bg-red-50 p-4 dark:bg-red-900/20">
              <p className="text-sm font-medium text-red-800 dark:text-red-400">
                ❌ Backend connection failed
              </p>
              <p className="mt-1 text-sm text-red-600 dark:text-red-500">
                {error}
              </p>
            </div>
          )}

          {!health && !error && !loading && (
            <p className="text-sm text-zinc-500">
              Click the button above to check backend health status.
            </p>
          )}
        </div>

        {/* Tech Stack */}
        <div className="mt-8 flex flex-wrap gap-2">
          {["Next.js", "TypeScript", "Zod", "Axios", "Tailwind CSS"].map(
            (tech) => (
              <span
                key={tech}
                className="rounded-full bg-zinc-100 px-3 py-1 text-sm text-zinc-700 dark:bg-zinc-800 dark:text-zinc-300"
              >
                {tech}
              </span>
            ),
          )}
        </div>
      </main>
    </div>
  );
}
