import type { Metadata } from "next";
import { redirect } from "next/navigation";
import { getSession } from "@/lib/auth/session";
import RegisterForm from "./RegisterForm";
import styles from "./RegisterPage.module.css";

export const metadata: Metadata = {
  title: "Регистрация - Movie Tracker",
  description: "Регистрация аккаунта Movie Tracker.",
  robots: {
    index: false,
    follow: false,
  },
};

const RegisterPage = async () => {
  const session = await getSession();

  if (session.isAuthenticated) {
    redirect("/");
  }

  return (
    <div className={styles.container}>
      <RegisterForm />
    </div>
  );
};

export default RegisterPage;
