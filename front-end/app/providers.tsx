"use client";

import { SessionProvider, signOut, useSession } from "next-auth/react";
import { useEffect } from "react";

function SessionErrorHandler() {
  const { data: session } = useSession();
  useEffect(() => {
    if ((session as any)?.error === "RefreshTokenError") {
      signOut({ callbackUrl: "/auth/login" });
    }
  }, [session]);
  return null;
}

export default function Providers({ children }: { children: React.ReactNode }) {
  return (
    <SessionProvider>
      <SessionErrorHandler />
      {children}
    </SessionProvider>
  );
}
