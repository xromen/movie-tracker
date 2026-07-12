"use client";

import { useRouter } from "next/navigation";
import { FormEvent, useState, useTransition } from "react";
import { register } from "@/lib/api/auth";
import styles from "./RegisterPage.module.css";

const REGISTER_ERROR_ID = "register-form-error";

const RegisterForm = () => {
  const router = useRouter();
  const [error, setError] = useState("");
  const [isPending, startTransition] = useTransition();

  const handleSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setError("");

    const formData = new FormData(event.currentTarget);
    const password = String(formData.get("password") ?? "");
    const repeatPassword = String(formData.get("repeatPassword") ?? "");

    if (password !== repeatPassword) {
      setError("Пароли не совпадают");
      return;
    }

    startTransition(async () => {
      try {
        await register({
          email: String(formData.get("email") ?? ""),
          username: String(formData.get("username") ?? ""),
          password,
        });
      } catch {
        setError("Этот email или имя пользователя уже используется");
        return;
      }

      router.push("/");
      router.refresh();
    });
  };

  return (
    <div className={styles.card}>
      <h3 className={styles.title}>Регистрация</h3>

      <form onSubmit={handleSubmit} className={styles.form} aria-describedby={error ? REGISTER_ERROR_ID : undefined}>
        <div className={styles.field}>
          <input
            id="register-email"
            className={styles.input}
            name="email"
            type="email"
            placeholder="Введите email"
            required
            autoComplete="email"
            aria-invalid={Boolean(error)}
          />
        </div>

        <div className={styles.field}>
          <input
            id="register-username"
            className={styles.input}
            name="username"
            placeholder="Введите имя пользователя"
            required
            minLength={3}
            autoComplete="username"
            aria-invalid={Boolean(error)}
          />
        </div>

        <div className={styles.field}>
          <input
            id="register-password"
            className={styles.input}
            name="password"
            type="password"
            placeholder="Пароль"
            required
            minLength={3}
            autoComplete="new-password"
            aria-invalid={Boolean(error)}
          />
        </div>

        <div className={styles.field}>
          <input
            id="register-repeat-password"
            className={styles.input}
            name="repeatPassword"
            type="password"
            placeholder="Повторите пароль"
            required
            autoComplete="new-password"
            aria-invalid={Boolean(error)}
          />
        </div>

        {error && <p id={REGISTER_ERROR_ID} className={styles.error} role="alert">{error}</p>}

        <button type="submit" className={styles.button} disabled={isPending}>
          {isPending ? "Регистрация..." : "Зарегистрироваться"}
        </button>
      </form>
    </div>
  );
};

export default RegisterForm;