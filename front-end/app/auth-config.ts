import NextAuth from "next-auth";
import Credentials from "next-auth/providers/credentials";

const API_URL = process.env.API_URL || "http://localhost:8080";

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

        if (!res.ok) return null;

        const data = await res.json();

        return {
          id: String(data.user_id ?? data.id ?? ""),
          name: data.username ?? username,
          email: data.email ?? "",
          accessToken: data.access_token,
          refreshToken: data.refresh_token,
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
      }
      return token;
    },
    async session({ session, token }) {
      (session as any).accessToken = token.accessToken;
      (session as any).refreshToken = token.refreshToken;
      (session as any).userId = token.userId;
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
