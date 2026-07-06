import { z } from "zod";

// ---------------------------------------------------------------------------
// Health check — matches Go backend /healthz response
// ---------------------------------------------------------------------------
export const HealthzResponseSchema = z.object({
  timestamp: z.string(),
});

export type HealthzResponse = z.infer<typeof HealthzResponseSchema>;
