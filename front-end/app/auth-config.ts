import NextAuth, { CredentialsSignin } from "next-auth";
import Credentials from "next-auth/providers/credentials";

class EmailNotVerifiedError extends CredentialsSignin {
  code = "EMAIL_NOT_VERIFIED";
}

const API_URL = process.env.API_URL || "http://localhost:8080";

function parseJwtPayload(token: string): Record<string, unknown> {
  try {
    const base64 = token.split(".")[1];
    const json = Buffer.from(base64, "base64").toString("utf-8");
    return JSON.parse(json);
  } catch {
    return {};
  }
}

export const { handlers, signIn, signOut, auth } = NextAuth({
  providers: [
    Credentials({
      credentials: {
        username: { label: "Username", type: "text" },
        password: { label: "Password", type: "password" },
      },
      async authorize(credentials) {
        const username = credentials?.username as string | undefined;
        const password = credentials?.password as string | undefined;
        if (!username || !password) return null;

        const res = await fetch(`${API_URL}/v1/auth/login`, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ username, password }),
        });

        if (!res.ok) {
          const body = await res.json().catch(() => null);
          if (body?.code === "EMAIL_NOT_VERIFIED") {
            throw new EmailNotVerifiedError();
          }
          return null;
        }

        const data = await res.json();

        const claims = parseJwtPayload(data.access_token ?? "");

        return {
          id: String(data.user_id ?? data.id ?? ""),
          name: data.username ?? username,
          email: data.email ?? "",
          accessToken: data.access_token,
          refreshToken: data.refresh_token,
          roles: (claims.roles as string[]) ?? ["viewer"],
        };
      },
    }),
  ],
  callbacks: {
    async jwt({ token, user }) {
      if (user) {
        token.accessToken = (user as any).accessToken;
        token.refreshToken = (user as any).refreshToken;
        token.userId = user.id;
        token.roles = (user as any).roles;
        return token;
      }

      // Check if access token is still valid
      const claims = parseJwtPayload((token.accessToken as string) ?? "");
      const exp = claims.exp as number | undefined;
      const now = Math.floor(Date.now() / 1000);

      if (exp && now < exp) {
        return token;
      }

      // Access token expired — try refresh
      try {
        const res = await fetch(`${API_URL}/v1/auth/refresh`, {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            Authorization: `Bearer ${token.accessToken}`,
          },
          body: JSON.stringify({ refresh_token: token.refreshToken }),
        });

        if (!res.ok) {
          return { ...token, error: "RefreshTokenError" as const };
        }

        const data = await res.json();
        const newClaims = parseJwtPayload(data.access_token ?? "");

        return {
          ...token,
          accessToken: data.access_token,
          refreshToken: data.refresh_token ?? token.refreshToken,
          roles: (newClaims.roles as string[]) ?? token.roles,
          error: undefined,
        };
      } catch {
        return { ...token, error: "RefreshTokenError" as const };
      }
    },
    async session({ session, token }) {
      (session as any).accessToken = token.accessToken;
      (session as any).refreshToken = token.refreshToken;
      (session as any).userId = token.userId;
      (session as any).roles = token.roles;
      (session as any).error = token.error;
      return session;
    },
  },
  pages: {
    signIn: "/auth/login",
  },
  session: {
    strategy: "jwt",
  },
});
