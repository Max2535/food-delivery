"use client";

import { useSession } from "next-auth/react";
import { useRouter } from "next/navigation";
import { useEffect, useState, FormEvent } from "react";

// ─── Types ───────────────────────────────────────────────────────
interface GroupUser {
  id: string;
  username: string;
  email: string;
}

interface Group {
  id: string;
  name: string;
  description: string;
  is_active: boolean;
  roles: string[];
  users: GroupUser[];
  created_at: string;
  updated_at: string;
}

const AVAILABLE_ROLES = [
  "admin",
  "manager",
  "chef",
  "cashier",
  "delivery",
  "inventory",
  "viewer",
];

// ─── Mock Data ───────────────────────────────────────────────────
const INITIAL_GROUPS: Group[] = [
  {
    id: "1",
    name: "ผู้ดูแลระบบ",
    description: "กลุ่มผู้ดูแลระบบทั้งหมด",
    is_active: true,
    roles: ["admin", "manager"],
    users: [
      { id: "u1", username: "admin1", email: "admin1@food.com" },
      { id: "u2", username: "admin2", email: "admin2@food.com" },
    ],
    created_at: "2026-03-01T00:00:00Z",
    updated_at: "2026-03-20T00:00:00Z",
  },
  {
    id: "2",
    name: "พนักงานครัว",
    description: "กลุ่มพนักงานประจำครัว",
    is_active: true,
    roles: ["chef", "viewer"],
    users: [
      { id: "u3", username: "chef1", email: "chef1@food.com" },
      { id: "u4", username: "chef2", email: "chef2@food.com" },
      { id: "u5", username: "chef3", email: "chef3@food.com" },
    ],
    created_at: "2026-03-05T00:00:00Z",
    updated_at: "2026-03-18T00:00:00Z",
  },
  {
    id: "3",
    name: "พนักงานส่งอาหาร",
    description: "กลุ่มพนักงานจัดส่ง",
    is_active: false,
    roles: ["delivery"],
    users: [{ id: "u6", username: "rider1", email: "rider1@food.com" }],
    created_at: "2026-03-10T00:00:00Z",
    updated_at: "2026-03-15T00:00:00Z",
  },
];

// ─── Component ───────────────────────────────────────────────────
export default function GroupsPage() {
  const { data: session, status } = useSession();
  const router = useRouter();

  const [groups, setGroups] = useState<Group[]>(INITIAL_GROUPS);
  const [search, setSearch] = useState("");
  const [filterStatus, setFilterStatus] = useState<"all" | "active" | "inactive">("all");

  // Modal state
  const [showModal, setShowModal] = useState(false);
  const [editingGroup, setEditingGroup] = useState<Group | null>(null);
  const [formName, setFormName] = useState("");
  const [formDescription, setFormDescription] = useState("");
  const [formRoles, setFormRoles] = useState<string[]>([]);

  // Detail / user management
  const [detailGroup, setDetailGroup] = useState<Group | null>(null);
  const [newUserName, setNewUserName] = useState("");
  const [newUserEmail, setNewUserEmail] = useState("");

  // Delete confirmation
  const [deletingId, setDeletingId] = useState<string | null>(null);

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

  // ─── Helpers ─────────────────────────────────────────────────
  const filtered = groups.filter((g) => {
    const matchSearch =
      g.name.toLowerCase().includes(search.toLowerCase()) ||
      g.description.toLowerCase().includes(search.toLowerCase());
    const matchStatus =
      filterStatus === "all"
        ? true
        : filterStatus === "active"
        ? g.is_active
        : !g.is_active;
    return matchSearch && matchStatus;
  });

  // ─── CRUD Handlers ──────────────────────────────────────────
  function openCreate() {
    setEditingGroup(null);
    setFormName("");
    setFormDescription("");
    setFormRoles([]);
    setShowModal(true);
  }

  function openEdit(group: Group) {
    setEditingGroup(group);
    setFormName(group.name);
    setFormDescription(group.description);
    setFormRoles([...group.roles]);
    setShowModal(true);
  }

  function handleSave(e: FormEvent) {
    e.preventDefault();
    if (!formName.trim()) return;

    if (editingGroup) {
      // Update
      setGroups((prev) =>
        prev.map((g) =>
          g.id === editingGroup.id
            ? {
                ...g,
                name: formName.trim(),
                description: formDescription.trim(),
                roles: formRoles,
                updated_at: new Date().toISOString(),
              }
            : g
        )
      );
    } else {
      // Create
      const newGroup: Group = {
        id: crypto.randomUUID(),
        name: formName.trim(),
        description: formDescription.trim(),
        is_active: true,
        roles: formRoles,
        users: [],
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
      };
      setGroups((prev) => [...prev, newGroup]);
    }
    setShowModal(false);
  }

  function toggleActive(id: string) {
    setGroups((prev) =>
      prev.map((g) =>
        g.id === id
          ? { ...g, is_active: !g.is_active, updated_at: new Date().toISOString() }
          : g
      )
    );
  }

  function confirmDelete(id: string) {
    setDeletingId(id);
  }

  function handleDelete() {
    if (deletingId) {
      setGroups((prev) => prev.filter((g) => g.id !== deletingId));
      if (detailGroup?.id === deletingId) setDetailGroup(null);
      setDeletingId(null);
    }
  }

  // ─── Role toggle ────────────────────────────────────────────
  function toggleRole(role: string) {
    setFormRoles((prev) =>
      prev.includes(role) ? prev.filter((r) => r !== role) : [...prev, role]
    );
  }

  // ─── User management in detail view ─────────────────────────
  function addUser(e: FormEvent) {
    e.preventDefault();
    if (!detailGroup || !newUserName.trim()) return;

    const newUser: GroupUser = {
      id: crypto.randomUUID(),
      username: newUserName.trim(),
      email: newUserEmail.trim(),
    };

    setGroups((prev) =>
      prev.map((g) =>
        g.id === detailGroup.id
          ? { ...g, users: [...g.users, newUser], updated_at: new Date().toISOString() }
          : g
      )
    );
    setDetailGroup((prev) =>
      prev ? { ...prev, users: [...prev.users, newUser] } : prev
    );
    setNewUserName("");
    setNewUserEmail("");
  }

  function removeUser(userId: string) {
    if (!detailGroup) return;

    setGroups((prev) =>
      prev.map((g) =>
        g.id === detailGroup.id
          ? {
              ...g,
              users: g.users.filter((u) => u.id !== userId),
              updated_at: new Date().toISOString(),
            }
          : g
      )
    );
    setDetailGroup((prev) =>
      prev ? { ...prev, users: prev.users.filter((u) => u.id !== userId) } : prev
    );
  }

  // ─── Inline role management for detail view ─────────────────
  function addRoleToGroup(role: string) {
    if (!detailGroup) return;
    setGroups((prev) =>
      prev.map((g) =>
        g.id === detailGroup.id && !g.roles.includes(role)
          ? { ...g, roles: [...g.roles, role], updated_at: new Date().toISOString() }
          : g
      )
    );
    setDetailGroup((prev) =>
      prev && !prev.roles.includes(role)
        ? { ...prev, roles: [...prev.roles, role] }
        : prev
    );
  }

  function removeRoleFromGroup(role: string) {
    if (!detailGroup) return;
    setGroups((prev) =>
      prev.map((g) =>
        g.id === detailGroup.id
          ? { ...g, roles: g.roles.filter((r) => r !== role), updated_at: new Date().toISOString() }
          : g
      )
    );
    setDetailGroup((prev) =>
      prev ? { ...prev, roles: prev.roles.filter((r) => r !== role) } : prev
    );
  }

  // ─── Render ─────────────────────────────────────────────────
  return (
    <div className="max-w-6xl mx-auto px-4 py-8">
      {/* Header */}
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold">จัดการกลุ่ม</h1>
          <p className="text-sm text-gray-500 mt-1">จัดการกลุ่มผู้ใช้งานและสิทธิ์การเข้าถึง</p>
        </div>
        <button
          onClick={openCreate}
          className="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 transition text-sm font-medium"
        >
          + สร้างกลุ่มใหม่
        </button>
      </div>

      {/* Filters */}
      <div className="flex flex-col sm:flex-row gap-3 mb-6">
        <input
          type="text"
          placeholder="ค้นหากลุ่ม..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="flex-1 border rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
        />
        <select
          value={filterStatus}
          onChange={(e) => setFilterStatus(e.target.value as "all" | "active" | "inactive")}
          className="border rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
        >
          <option value="all">ทั้งหมด</option>
          <option value="active">Active</option>
          <option value="inactive">Inactive</option>
        </select>
      </div>

      {/* Groups list + Detail panel */}
      <div className="flex flex-col lg:flex-row gap-6">
        {/* Groups Table */}
        <div className={`${detailGroup ? "lg:w-1/2" : "w-full"} transition-all`}>
          {filtered.length === 0 ? (
            <div className="bg-white rounded-lg shadow p-8 text-center text-gray-400">
              ไม่พบกลุ่ม
            </div>
          ) : (
            <div className="space-y-3">
              {filtered.map((group) => (
                <div
                  key={group.id}
                  className={`bg-white rounded-lg shadow p-4 border-l-4 ${
                    group.is_active ? "border-green-500" : "border-gray-300"
                  } ${
                    detailGroup?.id === group.id ? "ring-2 ring-blue-400" : ""
                  }`}
                >
                  <div className="flex items-start justify-between">
                    <div
                      className="flex-1 cursor-pointer"
                      onClick={() =>
                        setDetailGroup(
                          detailGroup?.id === group.id ? null : group
                        )
                      }
                    >
                      <div className="flex items-center gap-2">
                        <h3 className="font-semibold text-gray-800">
                          {group.name}
                        </h3>
                        <span
                          className={`text-xs px-2 py-0.5 rounded-full font-medium ${
                            group.is_active
                              ? "bg-green-100 text-green-700"
                              : "bg-gray-100 text-gray-500"
                          }`}
                        >
                          {group.is_active ? "Active" : "Inactive"}
                        </span>
                      </div>
                      <p className="text-sm text-gray-500 mt-1">
                        {group.description || "—"}
                      </p>
                      <div className="flex items-center gap-4 mt-2 text-xs text-gray-400">
                        <span>{group.users.length} สมาชิก</span>
                        <span>{group.roles.length} สิทธิ์</span>
                      </div>
                      {/* Roles badges */}
                      {group.roles.length > 0 && (
                        <div className="flex flex-wrap gap-1 mt-2">
                          {group.roles.map((role) => (
                            <span
                              key={role}
                              className="text-xs bg-blue-50 text-blue-700 px-2 py-0.5 rounded"
                            >
                              {role}
                            </span>
                          ))}
                        </div>
                      )}
                    </div>

                    {/* Actions */}
                    <div className="flex items-center gap-1 ml-3 shrink-0">
                      <button
                        onClick={() => toggleActive(group.id)}
                        title={group.is_active ? "Deactivate" : "Activate"}
                        className={`p-1.5 rounded text-xs font-medium transition ${
                          group.is_active
                            ? "text-yellow-600 hover:bg-yellow-50"
                            : "text-green-600 hover:bg-green-50"
                        }`}
                      >
                        {group.is_active ? "ปิด" : "เปิด"}
                      </button>
                      <button
                        onClick={() => openEdit(group)}
                        className="p-1.5 rounded text-blue-600 hover:bg-blue-50 text-xs font-medium transition"
                      >
                        แก้ไข
                      </button>
                      <button
                        onClick={() => confirmDelete(group.id)}
                        className="p-1.5 rounded text-red-600 hover:bg-red-50 text-xs font-medium transition"
                      >
                        ลบ
                      </button>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Detail Panel */}
        {detailGroup && (
          <div className="lg:w-1/2">
            <div className="bg-white rounded-lg shadow p-6 sticky top-4">
              <div className="flex items-center justify-between mb-4">
                <h2 className="text-lg font-bold text-gray-800">
                  {detailGroup.name}
                </h2>
                <button
                  onClick={() => setDetailGroup(null)}
                  className="text-gray-400 hover:text-gray-600 text-lg"
                >
                  ✕
                </button>
              </div>

              {/* Roles section */}
              <div className="mb-6">
                <h3 className="text-sm font-semibold text-gray-600 mb-2">
                  สิทธิ์ (Roles)
                </h3>
                <div className="flex flex-wrap gap-2 mb-3">
                  {detailGroup.roles.map((role) => (
                    <span
                      key={role}
                      className="inline-flex items-center gap-1 text-sm bg-blue-100 text-blue-800 px-2.5 py-1 rounded-full"
                    >
                      {role}
                      <button
                        onClick={() => removeRoleFromGroup(role)}
                        className="text-blue-500 hover:text-red-500 ml-0.5"
                      >
                        ×
                      </button>
                    </span>
                  ))}
                  {detailGroup.roles.length === 0 && (
                    <span className="text-sm text-gray-400">ไม่มีสิทธิ์</span>
                  )}
                </div>
                {/* Add role dropdown */}
                {AVAILABLE_ROLES.filter(
                  (r) => !detailGroup.roles.includes(r)
                ).length > 0 && (
                  <div className="flex items-center gap-2">
                    <select
                      id="add-role-select"
                      className="border rounded px-2 py-1 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                      defaultValue=""
                      onChange={(e) => {
                        if (e.target.value) {
                          addRoleToGroup(e.target.value);
                          e.target.value = "";
                        }
                      }}
                    >
                      <option value="" disabled>
                        + เพิ่มสิทธิ์...
                      </option>
                      {AVAILABLE_ROLES.filter(
                        (r) => !detailGroup.roles.includes(r)
                      ).map((role) => (
                        <option key={role} value={role}>
                          {role}
                        </option>
                      ))}
                    </select>
                  </div>
                )}
              </div>

              {/* Users section */}
              <div>
                <h3 className="text-sm font-semibold text-gray-600 mb-2">
                  สมาชิก ({detailGroup.users.length})
                </h3>

                {detailGroup.users.length === 0 ? (
                  <p className="text-sm text-gray-400 mb-3">ไม่มีสมาชิก</p>
                ) : (
                  <div className="space-y-2 mb-3 max-h-60 overflow-y-auto">
                    {detailGroup.users.map((user) => (
                      <div
                        key={user.id}
                        className="flex items-center justify-between bg-gray-50 rounded px-3 py-2"
                      >
                        <div>
                          <p className="text-sm font-medium text-gray-700">
                            {user.username}
                          </p>
                          <p className="text-xs text-gray-400">{user.email}</p>
                        </div>
                        <button
                          onClick={() => removeUser(user.id)}
                          className="text-xs text-red-500 hover:text-red-700"
                        >
                          นำออก
                        </button>
                      </div>
                    ))}
                  </div>
                )}

                {/* Add user form */}
                <form
                  onSubmit={addUser}
                  className="flex flex-col gap-2 border-t pt-3"
                >
                  <input
                    type="text"
                    placeholder="Username"
                    value={newUserName}
                    onChange={(e) => setNewUserName(e.target.value)}
                    className="border rounded px-2 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                    required
                  />
                  <input
                    type="email"
                    placeholder="Email"
                    value={newUserEmail}
                    onChange={(e) => setNewUserEmail(e.target.value)}
                    className="border rounded px-2 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                  />
                  <button
                    type="submit"
                    className="bg-gray-800 text-white text-sm px-3 py-1.5 rounded hover:bg-gray-900 transition"
                  >
                    + เพิ่มสมาชิก
                  </button>
                </form>
              </div>
            </div>
          </div>
        )}
      </div>

      {/* ─── Create/Edit Modal ─────────────────────────────────── */}
      {showModal && (
        <div className="fixed inset-0 bg-black/40 flex items-center justify-center z-50">
          <div className="bg-white rounded-xl shadow-lg w-full max-w-md mx-4 p-6">
            <h2 className="text-lg font-bold mb-4">
              {editingGroup ? "แก้ไขกลุ่ม" : "สร้างกลุ่มใหม่"}
            </h2>
            <form onSubmit={handleSave} className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  ชื่อกลุ่ม *
                </label>
                <input
                  type="text"
                  value={formName}
                  onChange={(e) => setFormName(e.target.value)}
                  className="w-full border rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="เช่น ผู้ดูแลระบบ"
                  required
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  รายละเอียด
                </label>
                <textarea
                  value={formDescription}
                  onChange={(e) => setFormDescription(e.target.value)}
                  className="w-full border rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                  rows={2}
                  placeholder="คำอธิบายกลุ่ม"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  สิทธิ์ (Roles)
                </label>
                <div className="flex flex-wrap gap-2">
                  {AVAILABLE_ROLES.map((role) => (
                    <button
                      key={role}
                      type="button"
                      onClick={() => toggleRole(role)}
                      className={`text-sm px-3 py-1 rounded-full border transition ${
                        formRoles.includes(role)
                          ? "bg-blue-600 text-white border-blue-600"
                          : "bg-white text-gray-600 border-gray-300 hover:border-blue-400"
                      }`}
                    >
                      {role}
                    </button>
                  ))}
                </div>
              </div>
              <div className="flex justify-end gap-2 pt-2">
                <button
                  type="button"
                  onClick={() => setShowModal(false)}
                  className="px-4 py-2 text-sm text-gray-600 hover:text-gray-800 transition"
                >
                  ยกเลิก
                </button>
                <button
                  type="submit"
                  className="bg-blue-600 text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-blue-700 transition"
                >
                  {editingGroup ? "บันทึก" : "สร้างกลุ่ม"}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* ─── Delete Confirmation Modal ─────────────────────────── */}
      {deletingId && (
        <div className="fixed inset-0 bg-black/40 flex items-center justify-center z-50">
          <div className="bg-white rounded-xl shadow-lg w-full max-w-sm mx-4 p-6 text-center">
            <h2 className="text-lg font-bold mb-2 text-red-600">ยืนยันการลบ</h2>
            <p className="text-sm text-gray-500 mb-4">
              คุณต้องการลบกลุ่มนี้ใช่หรือไม่? การดำเนินการนี้ไม่สามารถย้อนกลับได้
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
                ลบกลุ่ม
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
