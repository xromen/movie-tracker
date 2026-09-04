"use client";

import styles from "./Layout.module.css";
import { getBindingUrl } from "@/lib/api/telegram";

const TelegramButton = () => {
  const handleTelegramBinding = async () => {
    const bindingUrl = await getBindingUrl()

    console.log(bindingUrl)

    window.open(bindingUrl.Url, "_blank", "noopener,noreferrer");
  };

  return (
    <button type="button" onClick={handleTelegramBinding} className={styles.navLink}>
      Телега
    </button>
  );
};

export default TelegramButton;
