"use client";

import { useTranslation } from "react-i18next";
import { supportedLanguages, type Language } from "@/i18n/config";
import styles from "./LanguageSwitcher.module.css";

// 语言偏好持久化的 localStorage key
const LANG_STORAGE_KEY = "shop-app-lang";

export function LanguageSwitcher() {
  const { i18n, t } = useTranslation("common");

  // 切换语言并持久化到 localStorage
  function change(lng: Language) {
    void i18n.changeLanguage(lng);
    if (typeof window !== "undefined") {
      localStorage.setItem(LANG_STORAGE_KEY, lng);
    }
  }

  return (
    <div className={styles.switcher}>
      {supportedLanguages.map((lng) => (
        <button
          key={lng}
          className={`${styles.langBtn} ${i18n.language === lng ? styles.active : ""}`}
          onClick={() => change(lng)}
        >
          {t(`language.${lng}`)}
        </button>
      ))}
    </div>
  );
}
