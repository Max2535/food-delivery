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

interface Role {
  ID: number;
  name: string;
}

interface Group {
  id: string;
  name: string;
  description: string;
  is_active: boolean;
  role_ids: number[];
  users: GroupUser[];
  created_at: string;
  updated_at: string;
}

// ─── Component ───────────────────────────────────────────────────
export default function GroupsPage() {
  const { data: session, status } = useSession();
  const router = useRouter();

  const [availableRoles, setAvailableRoles] = useState<Role[]>([]);

  const [groups, setGroups] = useState<Group[]>([]);
  const [search, setSearch] = useState("");
  const [filterStatus, setFilterStatus] = useState<"all" | "active" | "inactive">("all");

  // Modal state
  const [showModal, setShowModal] = useState(false);
  const [editingGroup, setEditingGroup] = useState<Group | null>(null);
  const [formName, setFormName] = useState("");
  const [formDescription, setFormDescription] = useState("");
  const [formRoleIds, setFormRoleIds] = useState<number[]>([]);

  // Detail / user management
  const [detailGroup, setDetailGroup] = useState<Group | null>(null);
  const [newUserName, setNewUserName] = useState("");
  const [newUserEmail, setNewUserEmail] = useState("");

  // Delete confirmation
  const [deletingId, setDeletingId] = useState<string | null>(null);

  // Save loading state
  const [saving, setSaving] = useState(false);

  // ─── Helper: resolve role ID → name ─────────────────────────
  function roleName(id: number): string {
    return availableRoles.find((r) => r.ID === id)?.name ?? `role#${id}`;
  }

  useEffect(() => {
    if (status === "unauthenticated") {
      router.replace("/auth/login");
    }
  }, [status, router]);

  useEffect(() => {
    const accessToken = (session as any)?.accessToken;
    if (status !== "authenticated" || !accessToken) return;

    const controller = new AbortController();
    //get roles for form options and display
    async function fetchRoles() {
      try {
        const res = await fetch("/api/auth/roles", {
          headers: {
            Authorization: `Bearer ${accessToken}`,
          },
          signal: controller.signal,
        });

        if (!res.ok) {
          console.error("ไม่สามารถโหลด roles ได้", res.status);
          return;
        }

        const json = await res.json();
        const roles: Role[] = Array.isArray(json) ? json : json.roles;
        if (Array.isArray(roles)) {
          setAvailableRoles(roles);
        }
      } catch (error) {
        if ((error as any).name !== "AbortError") {
          console.error("fetch roles error", error);
        }
      }
    }

    fetchRoles();

    //get groups (with users and roles) for display and management
    async function fetchUserGroup() {
      try {
        const res = await fetch("/api/auth/groups", {
          headers: {
            Authorization: `Bearer ${accessToken}`,
          },
          signal: controller.signal,
        });

        if (!res.ok) {
          console.error("ไม่สามารถโหลด groups ได้", res.status);
          return;
        }

        const json = await res.json();
        const raw: any[] = Array.isArray(json) ? json : json.groups;
        if (Array.isArray(raw)) {
          const groups: Group[] = raw.map((g) => ({
            id: String(g.ID ?? g.id),
            name: g.name,
            description: g.description ?? "",
            is_active: g.is_active,
            role_ids: (g.roles ?? []).map((r: any) => r.ID),
            users: (g.users ?? []).map((u: any) => ({
              id: String(u.ID ?? u.id),
              username: u.username,
              email: u.email,
            })),
            created_at: g.created_at ?? "",
            updated_at: g.updated_at ?? "",
          }));
          setGroups(groups);
        }
      } catch (error) {
        if ((error as any).name !== "AbortError") {
          console.error("fetch groups error", error);
        }
      }
    }

    fetchUserGroup();

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
    setFormRoleIds([]);
    setShowModal(true);
  }

  function openEdit(group: Group) {
    setEditingGroup(group);
    setFormName(group.name);
    setFormDescription(group.description);
    setFormRoleIds([...group.role_ids]);
    setShowModal(true);
  }

  async function handleSave(e: FormEvent) {
    e.preventDefault();
    if (!formName.trim() || saving) return;

    const accessToken = (session as any)?.accessToken;
    if (!accessToken) return;

    if (editingGroup) {
      setSaving(true);
      try {
        const res = await fetch("/api/auth/groups", {
          method: "PUT",
          headers: {
            "Content-Type": "application/json",
            Authorization: `Bearer ${accessToken}`,
          },
          body: JSON.stringify({
            id: Number(editingGroup.id),
            name: formName.trim(),
            description: formDescription.trim(),
            is_active: editingGroup.is_active,
            role_ids: formRoleIds,
            users: (editingGroup.users ?? []).map((u: any) => ({
              id: Number(u.ID ?? u.id),
            }))
            , // Assuming backend can handle user IDs for group membership
          }),
        });

        if (!res.ok) {
          const err = await res.json().catch(() => null);
          console.error("แก้ไขกลุ่มไม่สำเร็จ", res.status, err);
          alert("แก้ไขกลุ่มไม่สำเร็จ: " + (err?.message ?? res.statusText));
          return;
        }

        const json = await res.json();
        const g = json.group ?? json;
        const editedId = String(g.ID ?? g.id);

        setGroups((prev) =>
          prev.map((existing) =>
            existing.id === editedId
              ? {
                ...existing,
                name: g.name,
                description: g.description ?? "",
                is_active: g.is_active,
                role_ids: (g.roles ?? []).map((r: any) => r.ID ?? r.id),
                users: (g.users ?? []).map((u: any) => ({
                  id: String(u.ID ?? u.id),
                  username: u.username,
                  email: u.email,
                })),
                updated_at: g.updated_at ?? new Date().toISOString(),
              }
              : existing
          )
        );
      } catch (err) {
        console.error("แก้ไขกลุ่มไม่สำเร็จ", err);
        alert("เกิดข้อผิดพลาดในการแก้ไขกลุ่ม");
        return;
      } finally {
        setSaving(false);
      }
    } else {
      setSaving(true);
      try {
        const res = await fetch("/api/auth/groups", {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            Authorization: `Bearer ${accessToken}`,
          },
          body: JSON.stringify({
            name: formName.trim(),
            description: formDescription.trim(),
            is_active: true,
            role_ids: formRoleIds,
          }),
        });

        if (!res.ok) {
          const err = await res.json().catch(() => null);
          console.error("สร้างกลุ่มไม่สำเร็จ", res.status, err);
          alert("สร้างกลุ่มไม่สำเร็จ: " + (err?.message ?? res.statusText));
          return;
        }

        const json = await res.json();
        const g = json.group ?? json;
        const newGroup: Group = {
          id: String(g.ID ?? g.id),
          name: g.name,
          description: g.description ?? "",
          is_active: g.is_active,
          role_ids: (g.roles ?? []).map((r: any) => r.ID ?? r.id),
          users: (g.users ?? []).map((u: any) => ({
            id: String(u.ID ?? u.id),
            username: u.username,
            email: u.email,
          })),
          created_at: g.created_at ?? "",
          updated_at: g.updated_at ?? "",
        };
        setGroups((prev) => [...prev, newGroup]);
      } catch (err) {
        console.error("สร้างกลุ่มไม่สำเร็จ", err);
        alert("เกิดข้อผิดพลาดในการสร้างกลุ่ม");
        return;
      } finally {
        setSaving(false);
      }
    }
    setShowModal(false);
  }

  async function toggleActive(id: string) {
    const accessToken = (session as any)?.accessToken;
    if (!accessToken) return;

    const group = groups.find((g) => g.id === id);
    if (!group) return;

    const newIsActive = !group.is_active;

    // Optimistic update
    setGroups((prev) =>
      prev.map((g) =>
        g.id === id
          ? { ...g, is_active: newIsActive, updated_at: new Date().toISOString() }
          : g
      )
    );

    try {
      const res = await fetch("/api/auth/groups", {
        method: "PUT",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${accessToken}`,
        },
        body: JSON.stringify({
          id: Number(id),
          name: group.name,
          description: group.description,
          is_active: newIsActive,
          role_ids: group.role_ids,
        }),
      });

      if (!res.ok) {
        // Revert on failure
        setGroups((prev) =>
          prev.map((g) =>
            g.id === id ? { ...g, is_active: group.is_active } : g
          )
        );
        const err = await res.json().catch(() => null);
        console.error("อัพเดทสถานะกลุ่มไม่สำเร็จ", res.status, err);
        alert("อัพเดทสถานะไม่สำเร็จ: " + (err?.message ?? res.statusText));
      }
    } catch (err) {
      // Revert on error
      setGroups((prev) =>
        prev.map((g) =>
          g.id === id ? { ...g, is_active: group.is_active } : g
        )
      );
      console.error("อัพเดทสถานะกลุ่มไม่สำเร็จ", err);
      alert("เกิดข้อผิดพลาดในการอัพเดทสถานะ");
    }
  }

  function confirmDelete(id: string) {
    setDeletingId(id);
  }

  async function handleDelete() {
    if (!deletingId) return;

    const accessToken = (session as any)?.accessToken;
    if (!accessToken) return;

    try {
      const res = await fetch("/api/auth/groups", {
        method: "DELETE",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${accessToken}`,
        },
        body: JSON.stringify({ id: Number(deletingId) }),
      });

      if (!res.ok) {
        const err = await res.json().catch(() => null);
        console.error("ลบกลุ่มไม่สำเร็จ", res.status, err);
        alert("ลบกลุ่มไม่สำเร็จ: " + (err?.message ?? res.statusText));
        return;
      }

      setGroups((prev) => prev.filter((g) => g.id !== deletingId));
      if (detailGroup?.id === deletingId) setDetailGroup(null);
    } catch (err) {
      console.error("ลบกลุ่มไม่สำเร็จ", err);
      alert("เกิดข้อผิดพลาดในการลบกลุ่ม");
    } finally {
      setDeletingId(null);
    }
  }

  // ─── Role toggle (modal) ───────────────────────────────────
  function toggleFormRole(roleId: number) {
    setFormRoleIds((prev) =>
      prev.includes(roleId) ? prev.filter((id) => id !== roleId) : [...prev, roleId]
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
  function addRoleToGroup(roleId: number) {
    if (!detailGroup) return;
    setGroups((prev) =>
      prev.map((g) =>
        g.id === detailGroup.id && !g.role_ids.includes(roleId)
          ? { ...g, role_ids: [...g.role_ids, roleId], updated_at: new Date().toISOString() }
          : g
      )
    );
    setDetailGroup((prev) =>
      prev && !prev.role_ids.includes(roleId)
        ? { ...prev, role_ids: [...prev.role_ids, roleId] }
        : prev
    );
  }

  function removeRoleFromGroup(roleId: number) {
    if (!detailGroup) return;
    setGroups((prev) =>
      prev.map((g) =>
        g.id === detailGroup.id
          ? { ...g, role_ids: g.role_ids.filter((id) => id !== roleId), updated_at: new Date().toISOString() }
          : g
      )
    );
    setDetailGroup((prev) =>
      prev ? { ...prev, role_ids: prev.role_ids.filter((id) => id !== roleId) } : prev
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
                  className={`bg-white rounded-lg shadow p-4 border-l-4 ${group.is_active ? "border-green-500" : "border-gray-300"
                    } ${detailGroup?.id === group.id ? "ring-2 ring-blue-400" : ""
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
                          className={`text-xs px-2 py-0.5 rounded-full font-medium ${group.is_active
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
                        <span>{group.role_ids.length} สิทธิ์</span>
                      </div>
                      {/* Roles badges */}
                      {group.role_ids.length > 0 && (
                        <div className="flex flex-wrap gap-1 mt-2">
                          {group.role_ids.map((id) => (
                            <span
                              key={id}
                              className="text-xs bg-blue-50 text-blue-700 px-2 py-0.5 rounded"
                            >
                              {roleName(id)}
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
                        className={`p-1.5 rounded text-xs font-medium transition ${group.is_active
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
                  {detailGroup.role_ids.map((id) => (
                    <span
                      key={id}
                      className="inline-flex items-center gap-1 text-sm bg-blue-100 text-blue-800 px-2.5 py-1 rounded-full"
                    >
                      {roleName(id)}
                      <button
                        onClick={() => removeRoleFromGroup(id)}
                        className="text-blue-500 hover:text-red-500 ml-0.5"
                      >
                        ×
                      </button>
                    </span>
                  ))}
                  {detailGroup.role_ids.length === 0 && (
                    <span className="text-sm text-gray-400">ไม่มีสิทธิ์</span>
                  )}
                </div>
                {/* Add role dropdown */}
                {availableRoles.filter(
                  (r) => !detailGroup.role_ids.includes(r.ID)
                ).length > 0 && (
                    <div className="flex items-center gap-2">
                      <select
                        id="add-role-select"
                        className="border rounded px-2 py-1 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                        defaultValue=""
                        onChange={(e) => {
                          if (e.target.value) {
                            addRoleToGroup(Number(e.target.value));
                            e.target.value = "";
                          }
                        }}
                      >
                        <option value="" disabled>
                          + เพิ่มสิทธิ์...
                        </option>
                        {availableRoles
                          .filter((r) => !detailGroup.role_ids.includes(r.ID))
                          .map((role) => (
                            <option key={role.ID} value={role.ID}>
                              {role.name}
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
                  {availableRoles.map((role) => (
                    <button
                      key={role.ID}
                      type="button"
                      onClick={() => toggleFormRole(role.ID)}
                      className={`text-sm px-3 py-1 rounded-full border transition ${formRoleIds.includes(role.ID)
                        ? "bg-blue-600 text-white border-blue-600"
                        : "bg-white text-gray-600 border-gray-300 hover:border-blue-400"
                        }`}
                    >
                      {role.name}
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
