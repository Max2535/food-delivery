"use client";

import { useSession } from "next-auth/react";
import { useRouter } from "next/navigation";
import { useEffect } from "react";
import Link from "next/link";
import { NAV_GROUPS, filterNavGroupsByRoles } from "../nav-config";

const SECTION_ICONS: Record<string, React.ReactNode> = {
  Auth: (
    <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
    </svg>
  ),
  Catalog: (
    <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253" />
    </svg>
  ),
  Kitchen: (
    <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M17.657 18.657A8 8 0 016.343 7.343S7 9 9 10c0-2 .5-5 2.986-7C14 5 16.09 5.777 17.656 7.343A7.975 7.975 0 0120 13a7.975 7.975 0 01-2.343 5.657z" />
    </svg>
  ),
  Order: (
    <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-3 7h3m-3 4h3m-6-4h.01M9 16h.01" />
    </svg>
  ),
  Inventory: (
    <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" />
    </svg>
  ),
};

const SECTION_COLORS: Record<string, { bg: string; text: string; border: string }> = {
  Auth: { bg: "bg-violet-50", text: "text-violet-600", border: "border-violet-100 hover:border-violet-200" },
  Catalog: { bg: "bg-blue-50", text: "text-blue-600", border: "border-blue-100 hover:border-blue-200" },
  Kitchen: { bg: "bg-orange-50", text: "text-orange-600", border: "border-orange-100 hover:border-orange-200" },
  Order: { bg: "bg-emerald-50", text: "text-emerald-600", border: "border-emerald-100 hover:border-emerald-200" },
  Inventory: { bg: "bg-amber-50", text: "text-amber-600", border: "border-amber-100 hover:border-amber-200" },
};

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
        <div className="w-8 h-8 border-2 border-primary-200 border-t-primary-600 rounded-full animate-spin" />
      </div>
    );
  }

  if (!session) return null;

  const userRoles: string[] = (session as any)?.roles ?? [];
  const visibleGroups = filterNavGroupsByRoles(NAV_GROUPS, userRoles);

  return (
    <div className="max-w-6xl mx-auto px-4 py-8 animate-fade-in">
      {/* Welcome */}
      <div className="bg-gradient-to-r from-primary-600 to-primary-700 rounded-2xl p-6 md:p-8 mb-8 text-white shadow-lg shadow-primary-600/20">
        <div className="flex items-center gap-4">
          <div className="w-14 h-14 bg-white/20 backdrop-blur rounded-2xl flex items-center justify-center text-2xl font-bold shrink-0">
            {(session.user?.name ?? "U")[0].toUpperCase()}
          </div>
          <div>
            <h1 className="text-xl md:text-2xl font-bold">
              สวัสดี, {session.user?.name}
            </h1>
            <p className="text-primary-100 text-sm mt-0.5">
              {session.user?.email}
              {(session as any)?.group && (
                <span className="ml-2 inline-flex items-center bg-white/15 backdrop-blur px-2 py-0.5 rounded-full text-xs">
                  {(session as any).group}
                </span>
              )}
            </p>
          </div>
        </div>
      </div>

      {/* Quick actions */}
      <h2 className="text-lg font-semibold text-gray-800 mb-4">เมนูลัด</h2>
      <div className="grid sm:grid-cols-2 lg:grid-cols-3 gap-4">
        {visibleGroups.map((group) => {
          const colors = SECTION_COLORS[group.label] ?? {
            bg: "bg-gray-50",
            text: "text-gray-600",
            border: "border-gray-100 hover:border-gray-200",
          };
          return (
            <div
              key={group.label}
              className={`bg-white rounded-xl border ${colors.border} p-5 transition-all hover:shadow-md`}
            >
              <div className="flex items-center gap-3 mb-4">
                <div className={`w-10 h-10 ${colors.bg} ${colors.text} rounded-xl flex items-center justify-center`}>
                  {SECTION_ICONS[group.label] ?? null}
                </div>
                <h3 className="font-semibold text-gray-800">{group.label}</h3>
              </div>
              <div className="space-y-1.5">
                {group.items.map((item) => (
                  <Link
                    key={item.href}
                    href={item.href}
                    className="flex items-center justify-between px-3 py-2 rounded-lg text-sm text-gray-600 hover:bg-gray-50 hover:text-gray-900 transition-colors group"
                  >
                    <span>{item.label}</span>
                    <svg
                      className="w-4 h-4 text-gray-300 group-hover:text-gray-500 transition-colors"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
                    </svg>
                  </Link>
                ))}
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}
