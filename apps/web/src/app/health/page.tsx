import type { Metadata } from "next";
import { redirect } from "next/navigation";
import { getSession } from "@/lib/auth/session";
import HealthClient from "./HealthClient";

export const metadata: Metadata = {
  title: "Состояние приложения - Movie Tracker",
  description: "Служебная страница состояния Movie Tracker.",
  robots: {
    index: false,
    follow: false,
  },
};

const HealthPage = async () => {
  const session = await getSession();
  const isAdmin = session.roles.some((role) => role.toLowerCase() === "admin");

  if (!session.isAuthenticated) {
    redirect("/login?from=/health");
  }

  if (!isAdmin) {
    redirect("/");
  }

  return <HealthClient />;
};

export default HealthPage;
