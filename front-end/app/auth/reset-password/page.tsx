"use client";

import Link from "next/link";
import { FormEvent, useState } from "react";
import { useSearchParams } from "next/navigation";
import { Suspense } from "react";

function ResetPasswordForm() {
  const searchParams = useSearchParams();
  const tokenFromUrl = searchParams.get("token") || "";

  const [token, setToken] = useState(tokenFromUrl);
  const [newPassword, setNewPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [loading, setLoading] = useState(false);

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setError("");
    setSuccess("");

    if (newPassword !== confirmPassword) {
      setError("รหัสผ่านไม่ตรงกัน");
      return;
    }

    if (newPassword.length < 8) {
      setError("รหัสผ่านต้องมีอย่างน้อย 8 ตัวอักษร");
      return;
    }

    setLoading(true);

    try {
      const res = await fetch("/api/auth/reset-password", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ token, new_password: newPassword }),
      });

      const data = await res.json().catch(() => null);

      if (!res.ok) {
        setError(data?.error || "ไม่สามารถรีเซ็ตรหัสผ่านได้");
        setLoading(false);
        return;
      }

      setSuccess("รีเซ็ตรหัสผ่านสำเร็จ! กรุณาเข้าสู่ระบบด้วยรหัสผ่านใหม่");
    } catch {
      setError("ไม่สามารถเชื่อมต่อเซิร์ฟเวอร์ได้");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="w-full max-w-md bg-white rounded-lg shadow p-8">
      <h1 className="text-2xl font-bold text-center mb-2">รีเซ็ตรหัสผ่าน</h1>
      <p className="text-center text-sm text-gray-500 mb-6">
        กรอกรหัสผ่านใหม่ของคุณ
      </p>

      {error && (
        <div className="bg-red-50 text-red-600 p-3 rounded mb-4 text-sm">
          {error}
        </div>
      )}

      {success && (
        <div className="bg-green-50 text-green-700 p-3 rounded mb-4 text-sm">
          {success}
          <p className="mt-2">
            <Link href="/auth/login" className="text-blue-600 hover:underline font-medium">
              เข้าสู่ระบบ
            </Link>
          </p>
        </div>
      )}

      {!success && (
        <form onSubmit={handleSubmit} className="space-y-4">
          {!tokenFromUrl && (
            <div>
              <label htmlFor="token" className="block text-sm font-medium text-gray-700 mb-1">
                โทเค็นรีเซ็ต
              </label>
              <input
                id="token"
                type="text"
                required
                value={token}
                onChange={(e) => setToken(e.target.value)}
                className="w-full border border-gray-300 rounded px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="reset token"
              />
            </div>
          )}

          <div>
            <label htmlFor="newPassword" className="block text-sm font-medium text-gray-700 mb-1">
              รหัสผ่านใหม่
            </label>
            <input
              id="newPassword"
              type="password"
              required
              value={newPassword}
              onChange={(e) => setNewPassword(e.target.value)}
              className="w-full border border-gray-300 rounded px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
              placeholder="new password (min 8 characters)"
            />
          </div>

          <div>
            <label htmlFor="confirmPassword" className="block text-sm font-medium text-gray-700 mb-1">
              ยืนยันรหัสผ่านใหม่
            </label>
            <input
              id="confirmPassword"
              type="password"
              required
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              className="w-full border border-gray-300 rounded px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
              placeholder="confirm new password"
            />
          </div>

          <button
            type="submit"
            disabled={loading}
            className="w-full bg-blue-600 text-white py-2 rounded hover:bg-blue-700 disabled:opacity-50 transition"
          >
            {loading ? "กำลังรีเซ็ต..." : "รีเซ็ตรหัสผ่าน"}
          </button>
        </form>
      )}

      <p className="text-center text-sm text-gray-500 mt-4">
        <Link href="/auth/login" className="text-blue-600 hover:underline">
          กลับไปหน้าเข้าสู่ระบบ
        </Link>
      </p>
    </div>
  );
}

export default function ResetPasswordPage() {
  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50">
      <Suspense fallback={<div className="text-gray-500">Loading...</div>}>
        <ResetPasswordForm />
      </Suspense>
    </div>
  );
}
