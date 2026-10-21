"use client";

import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import { apiFetch } from "@/lib/api";
import { clearToken } from "@/lib/auth";
import { Button } from "@/components/ui";

type MeResponse = {
  usuario: { email: string; nombre: string; rol: string };
  rol: string;
};

const navLinkBase: React.CSSProperties = {
  fontSize: "var(--text-body)",
  textDecoration: "none",
  padding: "6px 12px",
  borderRadius: "var(--radius-buttons)",
};

export function AppNav() {
  const pathname = usePathname();
  const router = useRouter();
  const [profile, setProfile] = useState<MeResponse | null>(null);

  useEffect(() => {
    let cancelled = false;
    apiFetch<MeResponse>("/api/v1/auth/me")
      .then((data) => {
        if (!cancelled) setProfile(data);
      })
      .catch(() => {
        if (!cancelled) setProfile(null);
      });
    return () => {
      cancelled = true;
    };
  }, []);

  async function handleLogout() {
    try {
      await apiFetch("/api/v1/auth/logout", { method: "POST" });
    } catch {
      /* session may already be invalid */
    }
    clearToken();
    router.replace("/login");
  }

  function linkStyle(href: string): React.CSSProperties {
    const active = pathname === href || pathname.startsWith(`${href}/`);
    return {
      ...navLinkBase,
      color: active ? "var(--color-ink-black)" : "var(--color-warm-gray)",
      background: active ? "var(--color-sky-wash)" : "transparent",
      fontWeight: active ? "var(--font-weight-medium)" : "var(--font-weight-regular)",
    };
  }

  return (
    <header
      style={{
        display: "flex",
        alignItems: "center",
        gap: "var(--spacing-16)",
        padding: "var(--spacing-12) var(--spacing-24)",
        background: "var(--color-pure-white)",
        borderBottom: "1px solid var(--color-stone-border)",
      }}
    >
      <Link
        href="/caja"
        style={{
          fontFamily: "var(--font-family-display)",
          fontSize: "var(--text-subheading)",
          color: "var(--color-ink-black)",
          textDecoration: "none",
          marginRight: "auto",
        }}
      >
        Flow<span className="highlight-span">Facture</span>
      </Link>
      <nav style={{ display: "flex", gap: "var(--spacing-8)", flexWrap: "wrap" }}>
        <Link href="/caja" style={linkStyle("/caja")}>
          Caja
        </Link>
        <Link href="/comprobantes" style={linkStyle("/comprobantes")}>
          Comprobantes
        </Link>
        {profile?.rol === "administrador" ? (
          <>
            <Link href="/admin/productos" style={linkStyle("/admin/productos")}>
              Productos
            </Link>
            <Link href="/admin/clientes" style={linkStyle("/admin/clientes")}>
              Clientes
            </Link>
            <Link href="/admin/series" style={linkStyle("/admin/series")}>
              Series
            </Link>
            <Link href="/admin/resumenes" style={linkStyle("/admin/resumenes")}>
              Resúmenes
            </Link>
            <Link href="/admin/reportes" style={linkStyle("/admin/reportes")}>
              Reportes
            </Link>
          </>
        ) : null}
      </nav>
      {profile ? (
        <span
          style={{
            fontSize: "var(--text-body)",
            color: "var(--color-warm-gray)",
            maxWidth: "220px",
            overflow: "hidden",
            textOverflow: "ellipsis",
            whiteSpace: "nowrap",
          }}
          title={profile.usuario.email}
        >
          {profile.usuario.nombre || profile.usuario.email}
          {profile.rol ? ` · ${profile.rol}` : ""}
        </span>
      ) : null}
      <Button variant="ghost" type="button" onClick={() => void handleLogout()}>
        Salir
      </Button>
    </header>
  );
}
