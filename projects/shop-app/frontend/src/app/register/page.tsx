"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { useTranslation } from "react-i18next";
import { useAuthStore } from "@/stores/auth";
import { makeRegisterSchema } from "@/schemas/auth";
import { LanguageSwitcher } from "@/components/LanguageSwitcher";
import styles from "./register.module.css";

export default function RegisterPage() {
  const { t } = useTranslation("common");
  const router = useRouter();
  const register = useAuthStore((s) => s.register);
  const [form, setForm] = useState({
    username: "",
    password: "",
    nickname: "",
    email: "",
    phone: "",
  });
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  // 处理注册表单提交：先用函数式 schema 做前端校验，再调用 store 的 register
  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    const parsed = makeRegisterSchema(t).safeParse(form);
    if (!parsed.success) {
      setError(parsed.error.issues[0]?.message ?? t("error.inputInvalid"));
      return;
    }
    setLoading(true);
    try {
      await register(form);
      router.push("/login");
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : t("error.registerFailed"));
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
        <h3 className={styles.title}>{t("register.title")}</h3>
        <form onSubmit={onSubmit}>
          <input className={styles.input} placeholder={t("register.username")} value={form.username} onChange={(e) => setForm({ ...form, username: e.target.value })} />
          <input className={styles.input} type="password" placeholder={t("register.password")} value={form.password} onChange={(e) => setForm({ ...form, password: e.target.value })} />
          <input className={styles.input} placeholder={t("register.nickname")} value={form.nickname} onChange={(e) => setForm({ ...form, nickname: e.target.value })} />
          <input className={styles.input} type="email" placeholder={t("register.email")} value={form.email} onChange={(e) => setForm({ ...form, email: e.target.value })} />
          <input className={styles.input} placeholder={t("register.phone")} value={form.phone} onChange={(e) => setForm({ ...form, phone: e.target.value })} />
          {error && <p className={styles.error}>{error}</p>}
          <button className={styles.registerBtn} type="submit" disabled={loading}>
            {loading ? t("register.submitting") : t("register.submit")}
          </button>
        </form>
        <p className={styles.loginLink}>
          {t("register.login")}
          <Link href="/login">{t("register.loginLink")}</Link>
        </p>
      </div>
      <LanguageSwitcher />
    </div>
  );
}
