import type { Metadata } from "next";
import Layout from "@/components/layout/Layout";
import "./globals.css";

export const metadata: Metadata = {
  title: {
    default: "Movie Tracker - фильмы и сериалы",
    template: "%s",
  },
  description: "Movie Tracker - поиск фильмов и сериалов, рейтинги, трейлеры, описания и рекомендации.",
  metadataBase: new URL(process.env.NEXT_PUBLIC_SITE_URL ?? "https://movietracker.ru"),
};

const RootLayout = ({ children }: Readonly<{ children: React.ReactNode }>) => {
  return (
    <html lang="ru">
      <body>
        <Layout>{children}</Layout>
      </body>
    </html>
  );
};

export default RootLayout;
