"use client";

import { signIn } from "next-auth/react";
import Link from "next/link";
import { FormEvent, useState } from "react";

export default function LoginPage() {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setError("");
    setLoading(true);

    const res = await signIn("credentials", {
      username,
      password,
      redirect: false,
    });

    setLoading(false);

    if (res?.error) {
      setError("ชื่อผู้ใช้หรือรหัสผ่านไม่ถูกต้อง");
      return;
    }

    window.location.href = "/dashboard";
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50">
      <div className="w-full max-w-md bg-white rounded-lg shadow p-8">
        <h1 className="text-2xl font-bold text-center mb-6">เข้าสู่ระบบ</h1>

        {error && (
          <div className="bg-red-50 text-red-600 p-3 rounded mb-4 text-sm">
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label htmlFor="username" className="block text-sm font-medium text-gray-700 mb-1">
              ชื่อผู้ใช้
            </label>
            <input
              id="username"
              type="text"
              required
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              className="w-full border border-gray-300 rounded px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
              placeholder="username"
            />
          </div>

          <div>
            <label htmlFor="password" className="block text-sm font-medium text-gray-700 mb-1">
              รหัสผ่าน
            </label>
            <input
              id="password"
              type="password"
              required
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              className="w-full border border-gray-300 rounded px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
              placeholder="password"
            />
          </div>

          <button
            type="submit"
            disabled={loading}
            className="w-full bg-blue-600 text-white py-2 rounded hover:bg-blue-700 disabled:opacity-50 transition"
          >
            {loading ? "กำลังเข้าสู่ระบบ..." : "เข้าสู่ระบบ"}
          </button>
        </form>

        <div className="text-center text-sm text-gray-500 mt-4 space-y-2">
          <p>
            <Link href="/auth/forgot-password" className="text-blue-600 hover:underline">
              ลืมรหัสผ่าน?
            </Link>
          </p>
          <p>
            ยังไม่มีบัญชี?{" "}
            <Link href="/auth/register" className="text-blue-600 hover:underline">
              สมัครสมาชิก
            </Link>
          </p>
        </div>
      </div>
    </div>
  );
}
