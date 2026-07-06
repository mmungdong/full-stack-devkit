"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { useTranslation } from "react-i18next";
import { useAuthStore } from "@/stores/auth";
import { LanguageSwitcher } from "@/components/LanguageSwitcher";
import styles from "./page.module.css";

export default function HomePage() {
  const { t } = useTranslation("common");
  const router = useRouter();
  const { token, user, logout, isAuthenticated } = useAuthStore();

  // 客户端软守卫：未登录跳转 /login（静态导出下无服务端中间件）
  useEffect(() => {
    if (!isAuthenticated()) {
      router.replace("/login");
    }
  }, [isAuthenticated, router]);

  if (!token) return null;

  return (
    <div className={styles.container}>
      <div className="starsec" />
      <div className="starthird" />
      <div className="starfourth" />
      <div className="starfifth" />

      <h1 className={styles.title}>{t("home.welcome", { name: user?.username ?? "" })}</h1>
      <p className={styles.text}>{t("home.loginSuccess")}</p>
      <button
        className={styles.logoutBtn}
        onClick={() => {
          logout();
          router.push("/login");
        }}
      >
        {t("home.logout")}
      </button>
      <LanguageSwitcher />
    </div>
  );
}
