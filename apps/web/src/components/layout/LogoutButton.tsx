"use client";

import { useRouter } from "next/navigation";
import { logout } from "@/lib/api/auth";
import styles from "./Layout.module.css";

const LogoutButton = () => {
  const router = useRouter();

  const handleLogout = async () => {
    await logout();

    router.push("/login");
    router.refresh();
  };

  return (
    <button type="button" onClick={handleLogout} className={styles.navLink}>
      Выйти
    </button>
  );
};

export default LogoutButton;
