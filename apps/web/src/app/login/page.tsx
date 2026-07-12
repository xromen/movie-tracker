import type { Metadata } from "next";
import { redirect } from "next/navigation";
import { getSession } from "@/lib/auth/session";
import LoginForm from "./LoginForm";
import styles from "./LoginPage.module.css";

export const metadata: Metadata = {
  title: "Вход - Movie Tracker",
  description: "Вход в личный аккаунт Movie Tracker.",
  robots: {
    index: false,
    follow: false,
  },
};

const LoginPage = async () => {
  const session = await getSession();

  if (session.isAuthenticated) {
    redirect("/");
  }

  return (
    <div className={styles.container}>
      <LoginForm />
    </div>
  );
};

export default LoginPage;
