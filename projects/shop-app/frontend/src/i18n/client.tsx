"use client";

import { useEffect, useState, type ReactNode } from "react";
import { I18nextProvider } from "react-i18next";
import i18n from "i18next";
import { initI18n, defaultLanguage, type Language } from "./config";
import { webGet } from "@/api";
import { configResponseSchema } from "@/schemas/config";

const LANG_STORAGE_KEY = "shop-app-lang";

function getStoredLang(): Language | null {
  if (typeof window === "undefined") return null;
  const v = localStorage.getItem(LANG_STORAGE_KEY);
  if (v === "zh" || v === "en") return v;
  return null;
}

async function fetchServerLang(): Promise<Language> {
  try {
    const { data, error } = await webGet("/config", configResponseSchema);
    if (error || !data) return defaultLanguage;
    return data.defaultLanguage === "en" ? "en" : "zh";
  } catch {
    return defaultLanguage;
  }
}

export function I18nBootstrap({ children }: { children: ReactNode }) {
  const [ready, setReady] = useState(i18n.isInitialized);

  useEffect(() => {
    if (i18n.isInitialized) {
      setReady(true);
      return;
    }
    (async () => {
      const stored = getStoredLang();
      const lng = stored ?? (await fetchServerLang());
      await initI18n(lng);
      setReady(true);
    })();
  }, []);

  if (!ready) {
    return <div style={{ minHeight: "100vh", background: "#181828" }} />;
  }
  return <I18nextProvider i18n={i18n}>{children}</I18nextProvider>;
}
