"use client";

import Link from "next/link";
import { useSession } from "next-auth/react";

export default function Home() {
  const { data: session } = useSession();

  if (session) return null;
  return (
    <div className="min-h-[80vh] flex flex-col items-center justify-center text-center px-4">
      <h1 className="text-4xl font-bold text-gray-900 mb-4">Food Delivery</h1>
      <p className="text-gray-600 mb-8 max-w-md">
        ระบบสั่งอาหารออนไลน์ สั่งง่าย ส่งไว
      </p>
      <div className="flex gap-4">
        <Link
          href="/auth/login"
          className="bg-blue-600 text-white px-6 py-2 rounded hover:bg-blue-700 transition"
        >
          เข้าสู่ระบบ
        </Link>
        <Link
          href="/auth/register"
          className="border border-blue-600 text-blue-600 px-6 py-2 rounded hover:bg-blue-50 transition"
        >
          สมัครสมาชิก
        </Link>
      </div>
    </div>
  );
}
