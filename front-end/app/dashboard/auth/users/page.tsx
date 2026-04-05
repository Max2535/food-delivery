"use client";

import { useSession } from "next-auth/react";
import { useRouter } from "next/navigation";
import { useEffect, useState, FormEvent } from "react";

// ─── Types ───────────────────────────────────────────────────────
interface Group {
  id: number;
  name: string;
}

interface User {
  id: number;
  username: string;
  email: string;
  group_id: number;
  group: string;
  is_verified: boolean;
  created_at: string;
}

// ─── Component ───────────────────────────────────────────────────
export default function UsersPage() {
  const { data: session, status } = useSession();
  const router = useRouter();

  const [users, setUsers] = useState<User[]>([]);
  const [groups, setGroups] = useState<Group[]>([]);
  const [search, setSearch] = useState("");
  const [filterGroup, setFilterGroup] = useState("all");
  const [filterVerified, setFilterVerified] = useState<"all" | "verified" | "unverified">("all");

  // Edit modal
  const [editingUser, setEditingUser] = useState<User | null>(null);
  const [formGroupId, setFormGroupId] = useState<number>(0);
  const [saving, setSaving] = useState(false);

  // Delete confirmation
  const [deletingId, setDeletingId] = useState<number | null>(null);

  useEffect(() => {
    if (status === "unauthenticated") router.replace("/auth/login");
  }, [status, router]);

  useEffect(() => {
    const token = (session as any)?.accessToken;
    if (status !== "authenticated" || !token) return;

    const controller = new AbortController();

    async function fetchAll() {
      try {
        const [usersRes, groupsRes] = await Promise.all([
          fetch("/api/auth/users", {
            headers: { Authorization: `Bearer ${token}` },
            signal: controller.signal,
          }),
          fetch("/api/auth/groups", {
            headers: { Authorization: `Bearer ${token}` },
            signal: controller.signal,
          }),
        ]);

        if (usersRes.ok) {
          const json = await usersRes.json();
          const raw: any[] = Array.isArray(json) ? json : json.users ?? [];
          setUsers(
            raw.map((u) => ({
              id: u.id ?? u.ID,
              username: u.username,
              email: u.email,
              group_id: u.group_id,
              group: u.group ?? "",
              is_verified: u.is_verified,
              created_at: u.created_at ?? "",
            }))
          );
        }

        if (groupsRes.ok) {
          const json = await groupsRes.json();
          const raw: any[] = Array.isArray(json) ? json : json.groups ?? [];
          setGroups(raw.map((g) => ({ id: g.ID ?? g.id, name: g.name })));
        }
      } catch (err: any) {
        if (err.name !== "AbortError") console.error(err);
      }
    }

    fetchAll();
    return () => controller.abort();
  }, [status, session]);

  if (status === "loading") {
    return (
      <div className="min-h-[80vh] flex items-center justify-center">
        <p className="text-gray-500">กำลังโหลด...</p>
      </div>
    );
  }
  if (!session) return null;

  const token = (session as any)?.accessToken;

  // ─── Derived data ──────────────────────────────────────────────
  const filtered = users.filter((u) => {
    const matchSearch =
      u.username.toLowerCase().includes(search.toLowerCase()) ||
      u.email.toLowerCase().includes(search.toLowerCase());
    const matchGroup = filterGroup === "all" || u.group === filterGroup;
    const matchVerified =
      filterVerified === "all"
        ? true
        : filterVerified === "verified"
          ? u.is_verified
          : !u.is_verified;
    return matchSearch && matchGroup && matchVerified;
  });

  const uniqueGroups = Array.from(new Set(users.map((u) => u.group))).filter(Boolean);

  // ─── Handlers ──────────────────────────────────────────────────
  function openEdit(user: User) {
    setEditingUser(user);
    setFormGroupId(user.group_id);
  }

  async function handleSave(e: FormEvent) {
    e.preventDefault();
    if (!editingUser || !formGroupId || saving) return;
    setSaving(true);

    try {
      const res = await fetch("/api/auth/users", {
        method: "PUT",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({ user_id: editingUser.id, group_id: formGroupId }),
      });

      if (!res.ok) {
        const err = await res.json().catch(() => null);
        alert("อัพเดทผู้ใช้ไม่สำเร็จ: " + (err?.message ?? res.statusText));
        return;
      }

      const newGroupName = groups.find((g) => g.id === formGroupId)?.name ?? editingUser.group;
      setUsers((prev) =>
        prev.map((u) =>
          u.id === editingUser.id
            ? { ...u, group_id: formGroupId, group: newGroupName }
            : u
        )
      );
      setEditingUser(null);
    } catch (err) {
      console.error(err);
      alert("เกิดข้อผิดพลาด");
    } finally {
      setSaving(false);
    }
  }

  async function handleDelete() {
    if (!deletingId) return;

    try {
      const res = await fetch("/api/auth/users", {
        method: "DELETE",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({ id: deletingId }),
      });

      if (!res.ok) {
        const err = await res.json().catch(() => null);
        alert("ลบผู้ใช้ไม่สำเร็จ: " + (err?.message ?? res.statusText));
        return;
      }

      setUsers((prev) => prev.filter((u) => u.id !== deletingId));
    } catch (err) {
      console.error(err);
      alert("เกิดข้อผิดพลาด");
    } finally {
      setDeletingId(null);
    }
  }

  function formatDate(iso: string) {
    if (!iso) return "—";
    return new Date(iso).toLocaleDateString("th-TH", {
      year: "numeric",
      month: "short",
      day: "numeric",
    });
  }

  // ─── Render ─────────────────────────────────────────────────────
  return (
    <div className="max-w-6xl mx-auto px-4 py-8">
      {/* Header */}
      <div className="mb-6">
        <h1 className="text-2xl font-bold">จัดการผู้ใช้</h1>
        <p className="text-sm text-gray-500 mt-1">ดู แก้ไขกลุ่ม และลบผู้ใช้งานในระบบ</p>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-3 gap-4 mb-6">
        {[
          { label: "ผู้ใช้ทั้งหมด", value: users.length, color: "blue" },
          { label: "ยืนยันอีเมลแล้ว", value: users.filter((u) => u.is_verified).length, color: "green" },
          { label: "ยังไม่ยืนยัน", value: users.filter((u) => !u.is_verified).length, color: "yellow" },
        ].map(({ label, value, color }) => (
          <div key={label} className="bg-white rounded-lg shadow p-4 text-center">
            <p className={`text-2xl font-bold text-${color}-600`}>{value}</p>
            <p className="text-xs text-gray-500 mt-1">{label}</p>
          </div>
        ))}
      </div>

      {/* Filters */}
      <div className="flex flex-col sm:flex-row gap-3 mb-6">
        <input
          type="text"
          placeholder="ค้นหาชื่อหรืออีเมล..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="flex-1 border rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
        />
        <select
          value={filterGroup}
          onChange={(e) => setFilterGroup(e.target.value)}
          className="border rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
        >
          <option value="all">ทุกกลุ่ม</option>
          {uniqueGroups.map((g) => (
            <option key={g} value={g}>{g}</option>
          ))}
        </select>
        <select
          value={filterVerified}
          onChange={(e) => setFilterVerified(e.target.value as "all" | "verified" | "unverified")}
          className="border rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
        >
          <option value="all">ทุกสถานะ</option>
          <option value="verified">ยืนยันแล้ว</option>
          <option value="unverified">ยังไม่ยืนยัน</option>
        </select>
      </div>

      {/* Table */}
      <div className="bg-white rounded-lg shadow overflow-hidden">
        {filtered.length === 0 ? (
          <div className="p-8 text-center text-gray-400">ไม่พบผู้ใช้</div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead className="bg-gray-50 border-b">
                <tr>
                  <th className="px-4 py-3 text-left font-medium text-gray-600">ID</th>
                  <th className="px-4 py-3 text-left font-medium text-gray-600">ชื่อผู้ใช้</th>
                  <th className="px-4 py-3 text-left font-medium text-gray-600">อีเมล</th>
                  <th className="px-4 py-3 text-left font-medium text-gray-600">กลุ่ม</th>
                  <th className="px-4 py-3 text-left font-medium text-gray-600">สถานะ</th>
                  <th className="px-4 py-3 text-left font-medium text-gray-600">สร้างเมื่อ</th>
                  <th className="px-4 py-3 text-right font-medium text-gray-600">จัดการ</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100">
                {filtered.map((user) => (
                  <tr key={user.id} className="hover:bg-gray-50 transition">
                    <td className="px-4 py-3 text-gray-400">#{user.id}</td>
                    <td className="px-4 py-3 font-medium text-gray-800">{user.username}</td>
                    <td className="px-4 py-3 text-gray-500">{user.email}</td>
                    <td className="px-4 py-3">
                      <span className="text-xs bg-blue-50 text-blue-700 px-2 py-0.5 rounded-full font-medium">
                        {user.group || "—"}
                      </span>
                    </td>
                    <td className="px-4 py-3">
                      <span
                        className={`text-xs px-2 py-0.5 rounded-full font-medium ${
                          user.is_verified
                            ? "bg-green-100 text-green-700"
                            : "bg-yellow-100 text-yellow-700"
                        }`}
                      >
                        {user.is_verified ? "ยืนยันแล้ว" : "รอยืนยัน"}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-gray-400">{formatDate(user.created_at)}</td>
                    <td className="px-4 py-3 text-right">
                      <div className="flex justify-end gap-1">
                        <button
                          onClick={() => openEdit(user)}
                          className="px-2.5 py-1 rounded text-blue-600 hover:bg-blue-50 text-xs font-medium transition"
                        >
                          แก้ไขกลุ่ม
                        </button>
                        <button
                          onClick={() => setDeletingId(user.id)}
                          className="px-2.5 py-1 rounded text-red-600 hover:bg-red-50 text-xs font-medium transition"
                        >
                          ลบ
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      <p className="text-xs text-gray-400 mt-3">
        แสดง {filtered.length} จาก {users.length} ผู้ใช้
      </p>

      {/* ─── Edit Group Modal ─────────────────────────────────── */}
      {editingUser && (
        <div className="fixed inset-0 bg-black/40 flex items-center justify-center z-50">
          <div className="bg-white rounded-xl shadow-lg w-full max-w-sm mx-4 p-6">
            <h2 className="text-lg font-bold mb-1">แก้ไขกลุ่มผู้ใช้</h2>
            <p className="text-sm text-gray-500 mb-4">
              <span className="font-medium text-gray-700">{editingUser.username}</span>{" "}
              — {editingUser.email}
            </p>
            <form onSubmit={handleSave} className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">กลุ่ม</label>
                <select
                  value={formGroupId}
                  onChange={(e) => setFormGroupId(Number(e.target.value))}
                  className="w-full border rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                  required
                >
                  <option value={0} disabled>
                    เลือกกลุ่ม...
                  </option>
                  {groups.map((g) => (
                    <option key={g.id} value={g.id}>
                      {g.name}
                    </option>
                  ))}
                </select>
              </div>
              <div className="flex justify-end gap-2 pt-1">
                <button
                  type="button"
                  onClick={() => setEditingUser(null)}
                  className="px-4 py-2 text-sm text-gray-600 hover:text-gray-800 transition"
                >
                  ยกเลิก
                </button>
                <button
                  type="submit"
                  disabled={saving || formGroupId === 0}
                  className="bg-blue-600 text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-blue-700 disabled:opacity-50 transition"
                >
                  {saving ? "กำลังบันทึก..." : "บันทึก"}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* ─── Delete Confirmation Modal ─────────────────────────── */}
      {deletingId !== null && (
        <div className="fixed inset-0 bg-black/40 flex items-center justify-center z-50">
          <div className="bg-white rounded-xl shadow-lg w-full max-w-sm mx-4 p-6 text-center">
            <h2 className="text-lg font-bold mb-2 text-red-600">ยืนยันการลบ</h2>
            <p className="text-sm text-gray-500 mb-1">
              คุณต้องการลบผู้ใช้{" "}
              <span className="font-semibold text-gray-700">
                &ldquo;{users.find((u) => u.id === deletingId)?.username}&rdquo;
              </span>{" "}
              ใช่หรือไม่?
            </p>
            <p className="text-xs text-gray-400 mb-4">
              การดำเนินการนี้ไม่สามารถย้อนกลับได้
            </p>
            <div className="flex justify-center gap-3">
              <button
                onClick={() => setDeletingId(null)}
                className="px-4 py-2 text-sm text-gray-600 hover:text-gray-800 transition"
              >
                ยกเลิก
              </button>
              <button
                onClick={handleDelete}
                className="bg-red-600 text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-red-700 transition"
              >
                ลบผู้ใช้
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
