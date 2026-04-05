"use client";

import { useSession } from "next-auth/react";
import { useRouter } from "next/navigation";
import { useEffect, useState, FormEvent } from "react";

// ─── Types ───────────────────────────────────────────────────────
interface Ingredient {
  id: number;
  name: string;
  unit: string;
  created_at: string;
  updated_at: string;
}

const COMMON_UNITS = ["g", "ml", "piece", "kg", "l", "tbsp", "tsp"];

// ─── Component ───────────────────────────────────────────────────
export default function IngredientsPage() {
  const { data: session, status } = useSession();
  const router = useRouter();

  const [ingredients, setIngredients] = useState<Ingredient[]>([]);
  const [search, setSearch] = useState("");
  const [filterUnit, setFilterUnit] = useState("all");

  // Create modal
  const [showModal, setShowModal] = useState(false);
  const [formName, setFormName] = useState("");
  const [formUnit, setFormUnit] = useState("g");
  const [formUnitCustom, setFormUnitCustom] = useState("");
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    if (status === "unauthenticated") router.replace("/auth/login");
  }, [status, router]);

  useEffect(() => {
    const token = (session as any)?.accessToken;
    if (status !== "authenticated" || !token) return;

    const controller = new AbortController();

    fetch("/api/catalog/ingredients", {
      headers: { Authorization: `Bearer ${token}` },
      signal: controller.signal,
    })
      .then((r) => r.json())
      .then((json) => setIngredients(Array.isArray(json) ? json : json.data ?? []))
      .catch((err) => { if (err.name !== "AbortError") console.error(err); });

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
  const units = ["all", ...Array.from(new Set(ingredients.map((i) => i.unit)))];
  const filtered = ingredients.filter((i) => {
    const matchSearch = i.name.toLowerCase().includes(search.toLowerCase());
    const matchUnit = filterUnit === "all" || i.unit === filterUnit;
    return matchSearch && matchUnit;
  });

  // ─── Group by unit for display ─────────────────────────────────
  const grouped: Record<string, Ingredient[]> = {};
  filtered.forEach((i) => {
    if (!grouped[i.unit]) grouped[i.unit] = [];
    grouped[i.unit].push(i);
  });

  // ─── Handlers ──────────────────────────────────────────────────
  function openCreate() {
    setFormName(""); setFormUnit("g"); setFormUnitCustom("");
    setShowModal(true);
  }

  async function handleSave(e: FormEvent) {
    e.preventDefault();
    if (!formName.trim() || saving) return;
    setSaving(true);

    const unit = formUnit === "__custom__" ? formUnitCustom.trim() : formUnit;
    if (!unit) { alert("กรุณาระบุหน่วย"); setSaving(false); return; }

    try {
      const res = await fetch("/api/catalog/ingredients", {
        method: "POST",
        headers: { "Content-Type": "application/json", Authorization: `Bearer ${token}` },
        body: JSON.stringify({ name: formName.trim(), unit }),
      });

      if (!res.ok) {
        const err = await res.json().catch(() => null);
        alert("เพิ่มวัตถุดิบไม่สำเร็จ: " + (err?.message ?? res.statusText));
        return;
      }

      const json = await res.json();
      const saved: Ingredient = json.data ?? json;
      setIngredients((prev) => [...prev, saved]);
      setShowModal(false);
    } catch (err) {
      console.error(err);
      alert("เกิดข้อผิดพลาด");
    } finally {
      setSaving(false);
    }
  }

  // ─── Render ─────────────────────────────────────────────────────
  return (
    <div className="max-w-4xl mx-auto px-4 py-8">
      {/* Header */}
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold">จัดการวัตถุดิบ</h1>
          <p className="text-sm text-gray-500 mt-1">วัตถุดิบที่ใช้ในสูตรอาหาร (BOM)</p>
        </div>
        <button
          onClick={openCreate}
          className="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 transition text-sm font-medium"
        >
          + เพิ่มวัตถุดิบ
        </button>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-2 gap-4 mb-6">
        <div className="bg-white rounded-lg shadow p-4 text-center">
          <p className="text-2xl font-bold text-blue-600">{ingredients.length}</p>
          <p className="text-xs text-gray-500 mt-1">วัตถุดิบทั้งหมด</p>
        </div>
        <div className="bg-white rounded-lg shadow p-4 text-center">
          <p className="text-2xl font-bold text-purple-600">
            {new Set(ingredients.map((i) => i.unit)).size}
          </p>
          <p className="text-xs text-gray-500 mt-1">หน่วยที่ใช้</p>
        </div>
      </div>

      {/* Filters */}
      <div className="flex flex-col sm:flex-row gap-3 mb-6">
        <input
          type="text"
          placeholder="ค้นหาวัตถุดิบ..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="flex-1 border rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
        />
        <select
          value={filterUnit}
          onChange={(e) => setFilterUnit(e.target.value)}
          className="border rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
        >
          {units.map((u) => (
            <option key={u} value={u}>{u === "all" ? "ทุกหน่วย" : u}</option>
          ))}
        </select>
      </div>

      {/* Ingredients list */}
      {filtered.length === 0 ? (
        <div className="bg-white rounded-lg shadow p-8 text-center text-gray-400">
          ไม่พบวัตถุดิบ
        </div>
      ) : filterUnit !== "all" ? (
        // Single unit — flat list
        <div className="bg-white rounded-lg shadow overflow-hidden">
          <table className="w-full text-sm">
            <thead className="bg-gray-50 border-b">
              <tr>
                <th className="text-left px-4 py-3 font-medium text-gray-600">#</th>
                <th className="text-left px-4 py-3 font-medium text-gray-600">ชื่อวัตถุดิบ</th>
                <th className="text-left px-4 py-3 font-medium text-gray-600">หน่วย</th>
              </tr>
            </thead>
            <tbody className="divide-y">
              {filtered.map((ing, idx) => (
                <tr key={ing.id} className="hover:bg-gray-50">
                  <td className="px-4 py-3 text-gray-400">{idx + 1}</td>
                  <td className="px-4 py-3 font-medium text-gray-800">{ing.name}</td>
                  <td className="px-4 py-3">
                    <span className="bg-purple-50 text-purple-700 text-xs px-2 py-0.5 rounded">
                      {ing.unit}
                    </span>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      ) : (
        // Grouped by unit
        <div className="space-y-4">
          {Object.entries(grouped).map(([unit, items]) => (
            <div key={unit} className="bg-white rounded-lg shadow overflow-hidden">
              <div className="bg-gray-50 border-b px-4 py-2 flex items-center justify-between">
                <span className="text-sm font-semibold text-gray-700">
                  หน่วย: <span className="text-purple-600">{unit}</span>
                </span>
                <span className="text-xs text-gray-400">{items.length} รายการ</span>
              </div>
              <div className="grid grid-cols-2 sm:grid-cols-3 gap-2 p-4">
                {items.map((ing) => (
                  <div
                    key={ing.id}
                    className="flex items-center gap-2 bg-gray-50 rounded-lg px-3 py-2"
                  >
                    <div className="w-2 h-2 rounded-full bg-purple-400 shrink-0" />
                    <span className="text-sm text-gray-700 truncate">{ing.name}</span>
                  </div>
                ))}
              </div>
            </div>
          ))}
        </div>
      )}

      {/* ─── Create Modal ─────────────────────────────────────── */}
      {showModal && (
        <div className="fixed inset-0 bg-black/40 flex items-center justify-center z-50">
          <div className="bg-white rounded-xl shadow-lg w-full max-w-sm mx-4 p-6">
            <h2 className="text-lg font-bold mb-4">เพิ่มวัตถุดิบใหม่</h2>
            <form onSubmit={handleSave} className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">ชื่อวัตถุดิบ *</label>
                <input
                  type="text"
                  value={formName}
                  onChange={(e) => setFormName(e.target.value)}
                  className="w-full border rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="เช่น หมูสับ"
                  required
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">หน่วย *</label>
                <select
                  value={formUnit}
                  onChange={(e) => setFormUnit(e.target.value)}
                  className="w-full border rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                >
                  {COMMON_UNITS.map((u) => (
                    <option key={u} value={u}>{u}</option>
                  ))}
                  <option value="__custom__">อื่นๆ (ระบุเอง)</option>
                </select>
                {formUnit === "__custom__" && (
                  <input
                    type="text"
                    value={formUnitCustom}
                    onChange={(e) => setFormUnitCustom(e.target.value)}
                    className="mt-2 w-full border rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                    placeholder="ระบุหน่วย..."
                    required
                  />
                )}
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
                  disabled={saving}
                  className="bg-blue-600 text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-blue-700 disabled:opacity-50 transition"
                >
                  เพิ่มวัตถุดิบ
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
