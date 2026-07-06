"use client";

import { create } from "zustand";
import { persist } from "zustand/middleware";
import { webPost } from "@/api";
import {
  loginResponseSchema,
  registerResponseSchema,
} from "@/schemas/auth";

interface AuthUser {
  userID: string;
  username: string;
}

interface AuthState {
  token: string | null;
  user: AuthUser | null;
  login: (username: string, password: string) => Promise<void>;
  register: (form: {
    username: string;
    password: string;
    nickname: string;
    email: string;
    phone: string;
  }) => Promise<string>;
  logout: () => void;
  isAuthenticated: () => boolean;
}

// 认证状态 store，token 持久化到 localStorage
export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      token: null,
      user: null,
      login: async (username, password) => {
        const { data, error } = await webPost(
          "/login",
          loginResponseSchema,
          { username, password },
        );
        if (error) throw error;
        set({ token: data!.token, user: { userID: "", username } });
      },
      register: async (form) => {
        const { data, error } = await webPost(
          "/v1/users",
          registerResponseSchema,
          form,
        );
        if (error) throw error;
        return data!.userID;
      },
      logout: () => set({ token: null, user: null }),
      isAuthenticated: () => get().token !== null,
    }),
    { name: "shop-app-auth" },
  ),
);
