"use client";

import { signOut, useSession } from "next-auth/react";
import Link from "next/link";

export default function Navbar() {
  const { data: session, status } = useSession();

  return (
    <nav className="bg-white shadow">
      <div className="max-w-5xl mx-auto px-4 py-3 flex items-center justify-between">
        <Link href="/" className="text-lg font-bold text-blue-600">
          Food Delivery
        </Link>

        <div className="flex items-center gap-3">
          {status === "loading" ? (
            <span className="text-sm text-gray-400">...</span>
          ) : session?.user ? (
            <>
              <Link href="/dashboard" className="text-sm text-gray-700 hover:text-blue-600">
                Dashboard
              </Link>
              <span className="text-sm text-gray-500">{session.user.name}</span>
              <button
                onClick={() => signOut({ callbackUrl: "/" })}
                className="text-sm text-red-600 hover:underline"
              >
                ออกจากระบบ
              </button>
            </>
          ) : (
            <>
              <Link
                href="/auth/login"
                className="text-sm text-blue-600 hover:underline"
              >
                เข้าสู่ระบบ
              </Link>
              <Link
                href="/auth/register"
                className="text-sm bg-blue-600 text-white px-3 py-1 rounded hover:bg-blue-700 transition"
              >
                สมัครสมาชิก
              </Link>
            </>
          )}
        </div>
      </div>
    </nav>
  );
}
