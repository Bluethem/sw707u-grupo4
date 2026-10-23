"use client";

import { useRouter } from "next/navigation";
import { useEffect } from "react";
import { apiFetch } from "@/lib/api";
import { homePathForRol, isAuthenticated } from "@/lib/auth";

export default function HomePage() {
  const router = useRouter();

  useEffect(() => {
    if (!isAuthenticated()) {
      router.replace("/login");
      return;
    }
    apiFetch<{ rol: string }>("/api/v1/auth/me")
      .then((me) => router.replace(homePathForRol(me.rol)))
      .catch(() => router.replace("/caja"));
  }, [router]);

  return null;
}
