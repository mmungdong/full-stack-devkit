import type { Metadata } from "next";
import "./globals.css";
import { I18nBootstrap } from "@/i18n/client";

export const metadata: Metadata = {
  title: "Shop App",
  description: "基于 OneX 技术栈的电商后端服务",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="zh" className="h-full antialiased">
      <body className="min-h-full flex flex-col font-sans">
        <I18nBootstrap>{children}</I18nBootstrap>
      </body>
    </html>
  );
}
