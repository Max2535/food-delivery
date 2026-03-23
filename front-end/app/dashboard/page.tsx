"use client";

import { useSession } from "next-auth/react";
import { useRouter } from "next/navigation";
import { useEffect } from "react";

export default function DashboardPage() {
  const { data: session, status } = useSession();
  const router = useRouter();

  useEffect(() => {
    if (status === "unauthenticated") {
      router.replace("/auth/login");
    }
  }, [status, router]);

  if (status === "loading") {
    return (
      <div className="min-h-[80vh] flex items-center justify-center">
        <p className="text-gray-500">กำลังโหลด...</p>
      </div>
    );
  }

  if (!session) return null;

  return (
    <div className="max-w-5xl mx-auto px-4 py-8">
      <h1 className="text-2xl font-bold mb-4">Dashboard</h1>
      <div className="bg-white rounded-lg shadow p-6">
        <p className="text-gray-700">
          สวัสดี, <span className="font-semibold">{session.user?.name}</span>
        </p>
        <p className="text-sm text-gray-500 mt-1">{session.user?.email}</p>
      </div>
    </div>
  );
}
