"use client";

import Link from "next/link";
import { usePathname, useSearchParams } from "next/navigation";
import type { ReactNode } from "react";

type LoginLinkProps = {
  children: ReactNode;
  className?: string;
};

const LoginLink = ({ children, className }: LoginLinkProps) => {
  const pathname = usePathname();
  const searchParams = useSearchParams();
  const queryString = searchParams.toString();
  const currentPath = queryString ? `${pathname}?${queryString}` : pathname;
  const href =
    currentPath === "/login" ? "/login" : `/login?from=${encodeURIComponent(currentPath)}`;

  return (
    <Link href={href} className={className}>
      {children}
    </Link>
  );
};

export default LoginLink;
