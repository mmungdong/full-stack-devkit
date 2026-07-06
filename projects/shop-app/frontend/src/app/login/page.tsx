"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { useTranslation } from "react-i18next";
import { useAuthStore } from "@/stores/auth";
import { makeLoginSchema } from "@/schemas/auth";
import { LanguageSwitcher } from "@/components/LanguageSwitcher";
import styles from "./login.module.css";

export default function LoginPage() {
  const { t } = useTranslation("common");
  const router = useRouter();
  const login = useAuthStore((s) => s.login);
  const [form, setForm] = useState({ username: "", password: "" });
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    const parsed = makeLoginSchema(t).safeParse(form);
    if (!parsed.success) {
      setError(parsed.error.issues[0]?.message ?? t("error.inputInvalid"));
      return;
    }
    setLoading(true);
    try {
      await login(form.username, form.password);
      router.push("/");
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : t("error.loginFailed"));
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className={styles.container}>
      <div className="starsec" />
      <div className="starthird" />
      <div className="starfourth" />
      <div className="starfifth" />

      <div className={styles.card}>
        <h3 className={styles.title}>{t("login.title")}</h3>
        <p className={styles.subtitle}>{t("login.subtitle")}</p>
        <form onSubmit={onSubmit}>
          <input
            className={styles.input}
            placeholder={t("login.username")}
            value={form.username}
            onChange={(e) => setForm({ ...form, username: e.target.value })}
          />
          <input
            className={styles.input}
            type="password"
            placeholder={t("login.password")}
            value={form.password}
            onChange={(e) => setForm({ ...form, password: e.target.value })}
          />
          {error && <p className={styles.error}>{error}</p>}
          <button className={styles.loginBtn} type="submit" disabled={loading}>
            {loading ? t("login.submitting") : t("login.submit")}
          </button>
        </form>
        <a className={styles.forgotLink} onClick={() => alert(t("error.comingSoon"))}>
          {t("login.forgotPassword")}
        </a>
        <div className={styles.socialList}>
          {["f", "g", "t", "d"].map((s) => (
            <button
              key={s}
              className={styles.socialBtn}
              onClick={() => alert(t("error.comingSoon"))}
              aria-label="social login"
            >
              {s.toUpperCase()}
            </button>
          ))}
        </div>
        <p className={styles.registerLink}>
          {t("login.register")}
          <Link href="/register">{t("login.registerLink")}</Link>
        </p>
      </div>
      <LanguageSwitcher />
    </div>
  );
}
