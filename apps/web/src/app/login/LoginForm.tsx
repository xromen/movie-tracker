"use client";

import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { FormEvent, useState, useTransition } from "react";
import { login } from "@/lib/api/auth";
import styles from "./LoginPage.module.css";

const LOGIN_ERROR_ID = "login-form-error";

const LoginForm = () => {
  const router = useRouter();
  const searchParams = useSearchParams();
  const [error, setError] = useState("");
  const [isPending, startTransition] = useTransition();

  const getReturnPath = () => {
    const from = searchParams.get("from");

    if (!from || !from.startsWith("/") || from.startsWith("//")) {
      return "/";
    }

    return from;
  };

  const handleSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setError("");

    const formData = new FormData(event.currentTarget);

    startTransition(async () => {
      try {
        await login({
          email: String(formData.get("email") ?? ""),
          password: String(formData.get("password") ?? ""),
        });
      } catch {
        setError("Неверный email или пароль");
        return;
      }

      router.push(getReturnPath());
      router.refresh();
    });
  };

  return (
    <div className={styles.card}>
      <h3 className={styles.title}>Вход</h3>

      <form onSubmit={handleSubmit} className={styles.form} aria-describedby={error ? LOGIN_ERROR_ID : undefined}>
        <div className={styles.field}>
          <label className={styles.label} htmlFor="login-email">Email</label>
          <input
            id="login-email"
            className={styles.input}
            name="email"
            placeholder="Введите email"
            required
            autoComplete="email"
            aria-invalid={Boolean(error)}
          />
        </div>

        <div className={styles.field}>
          <label className={styles.label} htmlFor="login-password">Пароль</label>
          <input
            id="login-password"
            className={styles.input}
            name="password"
            type="password"
            placeholder="Введите пароль"
            required
            minLength={3}
            autoComplete="current-password"
            aria-invalid={Boolean(error)}
          />
        </div>

        {error && <p id={LOGIN_ERROR_ID} className={styles.error} role="alert">{error}</p>}

        <button type="submit" className={styles.button} disabled={isPending}>
          {isPending ? "Вхожу..." : "Войти"}
        </button>
        <p className={styles.hint}>
          Нет аккаунта? <Link href="/register">Зарегистрироваться</Link>
        </p>
      </form>
    </div>
  );
};

export default LoginForm;