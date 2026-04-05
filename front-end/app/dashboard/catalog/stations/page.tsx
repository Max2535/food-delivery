"use client";

import { useSession } from "next-auth/react";
import { useRouter } from "next/navigation";
import { useEffect, useState, FormEvent } from "react";

// ─── Types ───────────────────────────────────────────────────────
interface KitchenStation {
  id: number;
  name: string;
  description: string;
  created_at: string;
  updated_at: string;
}

// ─── Component ───────────────────────────────────────────────────
export default function StationsPage() {
  const { data: session, status } = useSession();
  const router = useRouter();

  const [stations, setStations] = useState<KitchenStation[]>([]);
  const [search, setSearch] = useState("");

  // Create modal
  const [showModal, setShowModal] = useState(false);
  const [formName, setFormName] = useState("");
  const [formDescription, setFormDescription] = useState("");
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    if (status === "unauthenticated") router.replace("/auth/login");
  }, [status, router]);

  useEffect(() => {
    const token = (session as any)?.accessToken;
    if (status !== "authenticated" || !token) return;

    const controller = new AbortController();

    fetch("/api/catalog/stations", {
      headers: { Authorization: `Bearer ${token}` },
      signal: controller.signal,
    })
      .then((r) => r.json())
      .then((json) => setStations(Array.isArray(json) ? json : json.data ?? []))
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

  const filtered = stations.filter((s) =>
    s.name.toLowerCase().includes(search.toLowerCase()) ||
    (s.description ?? "").toLowerCase().includes(search.toLowerCase())
  );

  function openCreate() {
    setFormName(""); setFormDescription("");
    setShowModal(true);
  }

  async function handleSave(e: FormEvent) {
    e.preventDefault();
    if (!formName.trim() || saving) return;
    setSaving(true);

    try {
      const res = await fetch("/api/catalog/stations", {
        method: "POST",
        headers: { "Content-Type": "application/json", Authorization: `Bearer ${token}` },
        body: JSON.stringify({ name: formName.trim(), description: formDescription.trim() }),
      });

      if (!res.ok) {
        const err = await res.json().catch(() => null);
        alert("เพิ่มสถานีไม่สำเร็จ: " + (err?.message ?? res.statusText));
        return;
      }

      const json = await res.json();
      const saved: KitchenStation = json.data ?? json;
      setStations((prev) => [...prev, saved]);
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
          <h1 className="text-2xl font-bold">สถานีครัว (Kitchen Stations)</h1>
          <p className="text-sm text-gray-500 mt-1">จัดการสถานีครัวที่ใช้ในการเตรียมอาหาร</p>
        </div>
        <button
          onClick={openCreate}
          className="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 transition text-sm font-medium"
        >
          + เพิ่มสถานี
        </button>
      </div>

      {/* Stats */}
      <div className="bg-white rounded-lg shadow p-4 mb-6 flex items-center gap-4">
        <div className="w-12 h-12 bg-orange-100 rounded-xl flex items-center justify-center shrink-0">
          <svg className="w-6 h-6 text-orange-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M17.657 18.657A8 8 0 016.343 7.343S7 9 9 10c0-2 .5-5 2.986-7C14 5 16.09 5.777 17.656 7.343A7.975 7.975 0 0120 13a7.975 7.975 0 01-2.343 5.657z" />
          </svg>
        </div>
        <div>
          <p className="text-2xl font-bold text-gray-800">{stations.length}</p>
          <p className="text-sm text-gray-500">สถานีครัวทั้งหมด</p>
        </div>
      </div>

      {/* Search */}
      <div className="mb-6">
        <input
          type="text"
          placeholder="ค้นหาสถานี..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="w-full border rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
        />
      </div>

      {/* Stations grid */}
      {filtered.length === 0 ? (
        <div className="bg-white rounded-lg shadow p-8 text-center text-gray-400">
          ไม่พบสถานีครัว
        </div>
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
          {filtered.map((station) => (
            <div
              key={station.id}
              className="bg-white rounded-lg shadow p-5 border-t-4 border-orange-400 hover:shadow-md transition"
            >
              <div className="flex items-start justify-between mb-2">
                <div className="w-10 h-10 bg-orange-50 rounded-lg flex items-center justify-center">
                  <svg className="w-5 h-5 text-orange-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M17.657 18.657A8 8 0 016.343 7.343S7 9 9 10c0-2 .5-5 2.986-7C14 5 16.09 5.777 17.656 7.343A7.975 7.975 0 0120 13a7.975 7.975 0 01-2.343 5.657z" />
                  </svg>
                </div>
                <span className="text-xs text-gray-400">#{station.id}</span>
              </div>
              <h3 className="font-semibold text-gray-800 mt-2">{station.name}</h3>
              {station.description && (
                <p className="text-sm text-gray-500 mt-1">{station.description}</p>
              )}
              <p className="text-xs text-gray-300 mt-3">
                {new Date(station.created_at).toLocaleDateString("th-TH")}
              </p>
            </div>
          ))}

          {/* Add new card */}
          <button
            onClick={openCreate}
            className="bg-white rounded-lg shadow p-5 border-2 border-dashed border-gray-200 hover:border-blue-300 text-gray-400 hover:text-blue-500 transition flex flex-col items-center justify-center gap-2 min-h-[130px]"
          >
            <svg className="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M12 4v16m8-8H4" />
            </svg>
            <span className="text-sm font-medium">เพิ่มสถานีใหม่</span>
          </button>
        </div>
      )}

      {/* ─── Create Modal ─────────────────────────────────────── */}
      {showModal && (
        <div className="fixed inset-0 bg-black/40 flex items-center justify-center z-50">
          <div className="bg-white rounded-xl shadow-lg w-full max-w-sm mx-4 p-6">
            <h2 className="text-lg font-bold mb-4">เพิ่มสถานีครัวใหม่</h2>
            <form onSubmit={handleSave} className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">ชื่อสถานี *</label>
                <input
                  type="text"
                  value={formName}
                  onChange={(e) => setFormName(e.target.value)}
                  className="w-full border rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="เช่น Hot Kitchen, Cold Kitchen, Bar"
                  required
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">รายละเอียด</label>
                <textarea
                  value={formDescription}
                  onChange={(e) => setFormDescription(e.target.value)}
                  className="w-full border rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                  rows={2}
                  placeholder="คำอธิบายสถานี"
                />
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
                  เพิ่มสถานี
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
