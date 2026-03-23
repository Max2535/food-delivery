"use client";

import Link from "next/link";
import { FormEvent, useState } from "react";

export default function ForgotPasswordPage() {
  const [email, setEmail] = useState("");
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [loading, setLoading] = useState(false);

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setError("");
    setSuccess("");
    setLoading(true);

    try {
      const res = await fetch("/api/auth/forgot-password", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email }),
      });

      const data = await res.json().catch(() => null);

      if (!res.ok) {
        setError(data?.error || "ไม่สามารถดำเนินการได้");
        setLoading(false);
        return;
      }

      setSuccess("หากอีเมลนี้มีอยู่ในระบบ ลิงก์รีเซ็ตรหัสผ่านจะถูกส่งไป");

      // Dev mode: ถ้า backend ส่ง reset_token กลับมา แสดงลิงก์
      if (data?.reset_token) {
        setSuccess(
          `ลิงก์รีเซ็ตรหัสผ่าน (dev mode): /auth/reset-password?token=${data.reset_token}`
        );
      }
    } catch {
      setError("ไม่สามารถเชื่อมต่อเซิร์ฟเวอร์ได้");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50">
      <div className="w-full max-w-md bg-white rounded-lg shadow p-8">
        <h1 className="text-2xl font-bold text-center mb-2">ลืมรหัสผ่าน</h1>
        <p className="text-center text-sm text-gray-500 mb-6">
          กรอกอีเมลที่ใช้สมัครสมาชิก เพื่อรับลิงก์รีเซ็ตรหัสผ่าน
        </p>

        {error && (
          <div className="bg-red-50 text-red-600 p-3 rounded mb-4 text-sm">
            {error}
          </div>
        )}

        {success && (
          <div className="bg-green-50 text-green-700 p-3 rounded mb-4 text-sm">
            {success}
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label htmlFor="email" className="block text-sm font-medium text-gray-700 mb-1">
              อีเมล
            </label>
            <input
              id="email"
              type="email"
              required
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              className="w-full border border-gray-300 rounded px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
              placeholder="email@example.com"
            />
          </div>

          <button
            type="submit"
            disabled={loading}
            className="w-full bg-blue-600 text-white py-2 rounded hover:bg-blue-700 disabled:opacity-50 transition"
          >
            {loading ? "กำลังดำเนินการ..." : "ส่งลิงก์รีเซ็ตรหัสผ่าน"}
          </button>
        </form>

        <p className="text-center text-sm text-gray-500 mt-4">
          <Link href="/auth/login" className="text-blue-600 hover:underline">
            กลับไปหน้าเข้าสู่ระบบ
          </Link>
        </p>
      </div>
    </div>
  );
}
