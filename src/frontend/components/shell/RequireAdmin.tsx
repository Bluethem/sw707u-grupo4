"use client";

import { useRouter } from "next/navigation";
import { useEffect, useState, type ReactNode } from "react";
import { apiFetch } from "@/lib/api";
import { isAuthenticated } from "@/lib/auth";

type MeResponse = {
  usuario: { email: string; nombre: string; rol: string };
  rol: string;
};

export function RequireAdmin({ children }: { children: ReactNode }) {
  const router = useRouter();
  const [allowed, setAllowed] = useState(false);

  useEffect(() => {
    if (!isAuthenticated()) {
      router.replace("/login");
      return;
    }

    let cancelled = false;
    apiFetch<MeResponse>("/api/v1/auth/me")
      .then((data) => {
        if (cancelled) return;
        if (data.rol !== "administrador") {
          router.replace("/caja");
          return;
        }
        setAllowed(true);
      })
      .catch(() => {
        if (!cancelled) router.replace("/login");
      });

    return () => {
      cancelled = true;
    };
  }, [router]);

  if (!allowed) {
    return null;
  }

  return <>{children}</>;
}
