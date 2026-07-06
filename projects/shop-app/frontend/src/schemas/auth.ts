import { z } from "zod";

// 翻译函数类型（接收 key 返回文案）
type TFunc = (key: string) => string;

// 登录表单校验（函数式，message 走 i18n）
export const makeLoginSchema = (t: TFunc) =>
  z.object({
    username: z.string().min(3, t("validation.usernameMin")),
    password: z.string().min(8, t("validation.passwordMin")),
  });
export type LoginForm = { username: string; password: string };

// 注册表单校验（phone 必填，对齐后端）
export const makeRegisterSchema = (t: TFunc) =>
  z.object({
    username: z.string().min(3, t("validation.usernameMin")),
    password: z.string().min(8, t("validation.passwordMin")),
    nickname: z.string().min(1, t("validation.nicknameRequired")),
    email: z.string().email(t("validation.emailInvalid")),
    phone: z
      .string()
      .min(1, t("validation.phoneRequired"))
      .regex(/^1[3-9]\d{9}$/, t("validation.phoneInvalid")),
  });
export type RegisterForm = {
  username: string;
  password: string;
  nickname: string;
  email: string;
  phone: string;
};

// API 响应 schema（不变）
export const loginResponseSchema = z.object({
  token: z.string(),
  expireAt: z
    .object({ seconds: z.number().optional(), nanos: z.number().optional() })
    .optional(),
});
export type LoginResponse = z.infer<typeof loginResponseSchema>;

export const registerResponseSchema = z.object({
  userID: z.string(),
});
export type RegisterResponse = z.infer<typeof registerResponseSchema>;
