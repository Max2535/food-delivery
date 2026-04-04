"use client";

import { signIn } from "next-auth/react";
import Link from "next/link";
import { FormEvent, useState } from "react";

export default function LoginPage() {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);
  const [emailNotVerified, setEmailNotVerified] = useState(false);
  const [resendEmail, setResendEmail] = useState("");
  const [resendLoading, setResendLoading] = useState(false);
  const [resendMessage, setResendMessage] = useState("");

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setError("");
    setEmailNotVerified(false);
    setResendMessage("");
    setLoading(true);

    const res = await signIn("credentials", {
      username,
      password,
      redirect: false,
    });

    setLoading(false);

    if (res?.error) {
      if (res.error.includes("EMAIL_NOT_VERIFIED")) {
        setEmailNotVerified(true);
        return;
      }
      setError("ชื่อผู้ใช้หรือรหัสผ่านไม่ถูกต้อง");
      return;
    }

    window.location.href = "/dashboard";
  }

  async function handleResend(e: FormEvent) {
    e.preventDefault();
    setResendMessage("");
    setResendLoading(true);

    try {
      const res = await fetch("/api/auth/resend-verification", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email: resendEmail }),
      });
      if (res.ok) {
        setResendMessage("ส่งลิงก์ยืนยันใหม่แล้ว กรุณาตรวจสอบอีเมลของคุณ");
      } else {
        setResendMessage("ไม่พบอีเมลนี้ในระบบ กรุณาตรวจสอบอีกครั้ง");
      }
    } catch {
      setResendMessage("ไม่สามารถส่งอีเมลได้ กรุณาลองใหม่อีกครั้ง");
    } finally {
      setResendLoading(false);
    }
  }

  return (
    <div className="min-h-[calc(100vh-4rem)] flex items-center justify-center px-4 py-12 animate-fade-in">
      <div className="w-full max-w-md">
        {/* Header */}
        <div className="text-center mb-8">
          <div className="w-14 h-14 bg-gradient-to-br from-primary-500 to-primary-700 rounded-2xl flex items-center justify-center mx-auto mb-4 shadow-lg shadow-primary-600/20">
            <svg className="w-7 h-7 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
            </svg>
          </div>
          <h1 className="text-2xl font-bold text-gray-900">เข้าสู่ระบบ</h1>
          <p className="text-sm text-gray-500 mt-1">ยินดีต้อนรับกลับมา</p>
        </div>

        {/* Card */}
        <div className="bg-white rounded-2xl shadow-sm border border-gray-100 p-8">
          {emailNotVerified ? (
            /* Email not verified state */
            <div className="space-y-5">
              <div className="flex flex-col items-center text-center gap-3">
                <div className="w-14 h-14 bg-amber-100 rounded-full flex items-center justify-center">
                  <svg className="w-7 h-7 text-amber-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
                  </svg>
                </div>
                <div>
                  <p className="font-semibold text-gray-900">อีเมลยังไม่ได้รับการยืนยัน</p>
                  <p className="text-sm text-gray-500 mt-1">
                    กรุณายืนยันอีเมลก่อนเข้าสู่ระบบ ตรวจสอบกล่องจดหมายของคุณ<br />หรือขอรับลิงก์ยืนยันใหม่ด้านล่าง
                  </p>
                </div>
              </div>

              <Link
                href="/auth/verify-email"
                className="block w-full bg-primary-600 text-white py-2.5 rounded-xl font-medium hover:bg-primary-700 transition-colors shadow-sm text-center text-sm"
              >
                ไปยืนยันอีเมล
              </Link>

              <div className="relative">
                <div className="absolute inset-0 flex items-center">
                  <div className="w-full border-t border-gray-100" />
                </div>
                <div className="relative flex justify-center text-xs text-gray-400">
                  <span className="bg-white px-2">หรือขอรับลิงก์ยืนยันใหม่</span>
                </div>
              </div>

              <form onSubmit={handleResend} className="space-y-3">
                <input
                  type="email"
                  required
                  value={resendEmail}
                  onChange={(e) => setResendEmail(e.target.value)}
                  className="w-full border border-gray-200 rounded-xl px-4 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent transition-shadow"
                  placeholder="กรอกอีเมลของคุณ"
                />
                <button
                  type="submit"
                  disabled={resendLoading}
                  className="w-full border border-primary-600 text-primary-600 py-2.5 rounded-xl font-medium hover:bg-primary-50 disabled:opacity-50 transition-colors text-sm"
                >
                  {resendLoading ? "กำลังส่ง..." : "ส่งลิงก์ยืนยันใหม่"}
                </button>
                {resendMessage && (
                  <p className={`text-xs text-center ${resendMessage.includes("ส่งลิงก์") ? "text-green-600" : "text-red-500"}`}>
                    {resendMessage}
                  </p>
                )}
              </form>

              <p className="text-center text-sm">
                <button
                  onClick={() => { setEmailNotVerified(false); setResendMessage(""); }}
                  className="text-primary-600 hover:text-primary-700 font-medium"
                >
                  กลับไปหน้าเข้าสู่ระบบ
                </button>
              </p>
            </div>
          ) : (
            /* Normal login form */
            <>
              {error && (
                <div className="flex items-center gap-2 bg-red-50 text-red-600 p-3 rounded-xl mb-5 text-sm border border-red-100">
                  <svg className="w-4 h-4 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                  {error}
                </div>
              )}

              <form onSubmit={handleSubmit} className="space-y-5">
                <div>
                  <label htmlFor="username" className="block text-sm font-medium text-gray-700 mb-1.5">
                    ชื่อผู้ใช้
                  </label>
                  <input
                    id="username"
                    type="text"
                    required
                    value={username}
                    onChange={(e) => setUsername(e.target.value)}
                    className="w-full border border-gray-200 rounded-xl px-4 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent transition-shadow"
                    placeholder="username"
                  />
                </div>

                <div>
                  <label htmlFor="password" className="block text-sm font-medium text-gray-700 mb-1.5">
                    รหัสผ่าน
                  </label>
                  <input
                    id="password"
                    type="password"
                    required
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    className="w-full border border-gray-200 rounded-xl px-4 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent transition-shadow"
                    placeholder="password"
                  />
                </div>

                <button
                  type="submit"
                  disabled={loading}
                  className="w-full bg-primary-600 text-white py-2.5 rounded-xl font-medium hover:bg-primary-700 disabled:opacity-50 transition-colors shadow-sm"
                >
                  {loading ? (
                    <span className="inline-flex items-center gap-2">
                      <svg className="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24">
                        <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
                        <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                      </svg>
                      กำลังเข้าสู่ระบบ...
                    </span>
                  ) : (
                    "เข้าสู่ระบบ"
                  )}
                </button>
              </form>

              <div className="mt-6 pt-5 border-t border-gray-100 text-center text-sm space-y-2">
                <p>
                  <Link href="/auth/forgot-password" className="text-primary-600 hover:text-primary-700 font-medium">
                    ลืมรหัสผ่าน?
                  </Link>
                </p>
                <p className="text-gray-500">
                  ยังไม่มีบัญชี?{" "}
                  <Link href="/auth/register" className="text-primary-600 hover:text-primary-700 font-medium">
                    สมัครสมาชิก
                  </Link>
                </p>
              </div>
            </>
          )}
        </div>
      </div>
    </div>
  );
}
