import { z } from "zod";

// /api/config 响应 schema（已确认返回纯数据 {"defaultLanguage":"zh"}，无信封）
export const configResponseSchema = z.object({
  defaultLanguage: z.string(),
});
