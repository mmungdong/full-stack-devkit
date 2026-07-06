import i18n from "i18next";
import { initReactI18next } from "react-i18next";
import zh from "@/locales/zh/common.json";
import en from "@/locales/en/common.json";

// 支持的语言列表（新增语言只需在此注册 + 加 locales/{lang}/common.json）
export const supportedLanguages = ["zh", "en"] as const;
export type Language = (typeof supportedLanguages)[number];

export const defaultLanguage: Language = "zh";

// 初始化 i18next。defaultLng 为初始语言（由调用方按优先级决定）.
export function initI18n(defaultLng: string) {
  if (i18n.isInitialized) {
    void i18n.changeLanguage(defaultLng);
    return i18n;
  }
  return i18n.use(initReactI18next).init({
    resources: { zh: { common: zh }, en: { common: en } },
    lng: defaultLng,
    fallbackLng: defaultLanguage,
    defaultNS: "common",
    ns: ["common"],
    interpolation: { escapeValue: false },
  });
}
