"use client";

import { AppNav } from "@/components/shell/AppNav";
import { RequireAdmin } from "@/components/shell/RequireAdmin";
import { RequireAuth } from "@/components/shell/RequireAuth";

export default function AdminLayout({ children }: { children: React.ReactNode }) {
  return (
    <RequireAuth>
      <RequireAdmin>
        <AppNav />
        <main
          style={{
            display: "block",
            width: "100%",
            minHeight: "calc(100vh - 57px)",
            padding: "var(--spacing-16)",
            maxWidth: "var(--page-max-width)",
            margin: "0 auto",
            textAlign: "left",
          }}
        >
          {children}
        </main>
      </RequireAdmin>
    </RequireAuth>
  );
}
