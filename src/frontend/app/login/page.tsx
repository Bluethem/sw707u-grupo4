"use client";

import { useRouter } from "next/navigation";
import { FormEvent, useEffect, useState } from "react";
import { Button, Card, Input } from "@/components/ui";
import { ApiError, apiFetch } from "@/lib/api";
import { homePathForRol, isAuthenticated, setToken } from "@/lib/auth";

export default function LoginPage() {
  const router = useRouter();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [formError, setFormError] = useState<string | undefined>();
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    if (!isAuthenticated()) return;
    apiFetch<{ rol: string }>("/api/v1/auth/me")
      .then((me) => router.replace(homePathForRol(me.rol)))
      .catch(() => router.replace("/caja"));
  }, [router]);

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    setFormError(undefined);
    setSubmitting(true);
    try {
      const data = await apiFetch<{ token: string }>("/api/v1/auth/login", {
        method: "POST",
        body: { email, password },
        token: null,
      });
      setToken(data.token);
      const me = await apiFetch<{ rol: string }>("/api/v1/auth/me");
      router.replace(homePathForRol(me.rol));
    } catch (err) {
      if (err instanceof ApiError && err.status === 401) {
        setFormError("Credenciales inválidas");
      } else if (err instanceof ApiError) {
        setFormError(err.message);
      } else {
        setFormError("No se pudo iniciar sesión");
      }
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <main className="auth-main">
      <Card style={{ width: "100%", maxWidth: "400px" }}>
        <h1
          style={{
            fontSize: "var(--text-heading-sm)",
            lineHeight: "var(--leading-heading-sm)",
            letterSpacing: "var(--tracking-heading-sm)",
            marginBottom: "var(--spacing-8)",
            textAlign: "center",
          }}
        >
          Inicia en tu caja <span className="highlight-span">rápida</span>
        </h1>
        <p
          style={{
            textAlign: "center",
            marginBottom: "var(--spacing-24)",
            fontSize: "var(--text-body)",
          }}
        >
          FlowFacture — punto de venta
        </p>
        <form
          onSubmit={(e) => void onSubmit(e)}
          style={{ display: "flex", flexDirection: "column", gap: "var(--spacing-16)" }}
        >
          <Input
            label="Correo"
            type="email"
            name="email"
            autoComplete="username"
            value={email}
            onChange={(ev) => setEmail(ev.target.value)}
            required
          />
          <Input
            label="Contraseña"
            type="password"
            name="password"
            autoComplete="current-password"
            value={password}
            onChange={(ev) => setPassword(ev.target.value)}
            required
          />
          {formError ? (
            <p style={{ margin: 0, fontSize: "var(--text-body)", color: "var(--color-error-text)" }}>
              {formError}
            </p>
          ) : null}
          <Button type="submit" disabled={submitting} style={{ width: "100%" }}>
            {submitting ? "Ingresando…" : "Iniciar sesión"}
          </Button>
        </form>
      </Card>
    </main>
  );
}
