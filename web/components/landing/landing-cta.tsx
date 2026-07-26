"use client";

import type { ReactNode } from "react";
import Link from "next/link";
import { trackEvent } from "@/lib/analytics";

export function SignupCta({
  className,
  children,
}: {
  className?: string;
  children: ReactNode;
}) {
  return (
    <Link
      href="/login?signup=1"
      className={className}
      data-analytics="signup-click"
      onClick={() => trackEvent("signup-click", { surface: "landing" })}
    >
      {children}
    </Link>
  );
}

export function InstallCta({ className }: { className?: string }) {
  return (
    <a
      href="https://github.com/cyberskill-official/shopass/tree/main/extension"
      className={className}
      data-analytics="install-click"
      onClick={() => trackEvent("install-click", { surface: "landing" })}
      target="_blank"
      rel="noreferrer"
    >
      Cài extension (Chrome)
    </a>
  );
}
