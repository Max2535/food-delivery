import { auth } from "@/app/auth-config";

export default auth((req) => {
  if (!req.auth && req.nextUrl.pathname.startsWith("/dashboard")) {
    const loginUrl = new URL("/auth/login", req.nextUrl.origin);
    return Response.redirect(loginUrl);
  }
});

export const config = {
  matcher: ["/dashboard/:path*"],
};
