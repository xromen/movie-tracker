"use client";

import { useRouter } from "next/navigation";
import { useEffect } from "react";

const AUTH_SESSION_CHANGED_EVENT = "movie-tracker:auth-session-changed";

const AuthSessionListener = () => {
  const router = useRouter();

  useEffect(() => {
    const handleAuthSessionChanged = () => {
      router.refresh();
    };

    window.addEventListener(AUTH_SESSION_CHANGED_EVENT, handleAuthSessionChanged);

    return () => {
      window.removeEventListener(AUTH_SESSION_CHANGED_EVENT, handleAuthSessionChanged);
    };
  }, [router]);

  return null;
};

export default AuthSessionListener;
