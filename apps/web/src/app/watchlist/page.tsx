import type { Metadata } from "next";
import { redirect } from "next/navigation";
import { getSession } from "@/lib/auth/session";
import WatchlistClient from "./WatchlistClient";

export const metadata: Metadata = {
  title: "Мой список - Movie Tracker",
  description: "Личный список фильмов и сериалов Movie Tracker.",
  robots: {
    index: false,
    follow: false,
  },
};

const WatchlistPage = async () => {
  const session = await getSession();

  if (!session.isAuthenticated) {
    redirect("/login?from=/watchlist");
  }

  return <WatchlistClient />;
};

export default WatchlistPage;
