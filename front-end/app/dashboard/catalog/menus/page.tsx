"use client";

import { useSession } from "next-auth/react";
import { useRouter } from "next/navigation";
import { useEffect, useState, FormEvent } from "react";

// ─── Types ───────────────────────────────────────────────────────
interface Ingredient {
  id: number;
  name: string;
  unit: string;
}

interface BOMItem {
  id: number;
  menu_item_id: number;
  ingredient_id?: number;
  sub_menu_item_id?: number;
  quantity: number;
  ingredient?: Ingredient;
  sub_menu_item?: { id: number; name: string };
}

interface MenuItem {
  id: number;
  name: string;
  description: string;
  price: number;
  category: string;
  is_available: boolean;
  created_at: string;
  updated_at: string;
}

// ─── Component ───────────────────────────────────────────────────
export default function MenusPage() {
  const { data: session, status } = useSession();
  const router = useRouter();

  const [menus, setMenus] = useState<MenuItem[]>([]);
  const [ingredients, setIngredients] = useState<Ingredient[]>([]);
  const [search, setSearch] = useState("");
  const [filterCategory, setFilterCategory] = useState("all");

  // Create/Edit modal
  const [showModal, setShowModal] = useState(false);
  const [editingMenu, setEditingMenu] = useState<MenuItem | null>(null);
  const [formName, setFormName] = useState("");
  const [formDescription, setFormDescription] = useState("");
  const [formPrice, setFormPrice] = useState("");
  const [formCategory, setFormCategory] = useState("");
  const [formAvailable, setFormAvailable] = useState(true);
  const [saving, setSaving] = useState(false);

  // BOM detail panel
  const [selectedMenu, setSelectedMenu] = useState<MenuItem | null>(null);
  const [bom, setBom] = useState<BOMItem[]>([]);
  const [bomLoading, setBomLoading] = useState(false);
  const [showAddBom, setShowAddBom] = useState(false);
  const [bomIngredientId, setBomIngredientId] = useState("");
  const [bomQuantity, setBomQuantity] = useState("");
  const [addingBom, setAddingBom] = useState(false);

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
        const [menusRes, ingredientsRes] = await Promise.all([
          fetch("/api/catalog/menus", {
            headers: { Authorization: `Bearer ${token}` },
            signal: controller.signal,
          }),
          fetch("/api/catalog/ingredients", {
            headers: { Authorization: `Bearer ${token}` },
            signal: controller.signal,
          }),
        ]);

        if (menusRes.ok) {
          const json = await menusRes.json();
          setMenus(Array.isArray(json) ? json : json.data ?? []);
        }
        if (ingredientsRes.ok) {
          const json = await ingredientsRes.json();
          setIngredients(Array.isArray(json) ? json : json.data ?? []);
        }
      } catch (err: any) {
        if (err.name !== "AbortError") console.error(err);
      }
    }

    fetchAll();
    return () => controller.abort();
  }, [status, session]);

  // Fetch BOM when a menu is selected
  useEffect(() => {
    if (!selectedMenu) { setBom([]); return; }
    const token = (session as any)?.accessToken;
    if (!token) return;

    setBomLoading(true);
    fetch(`/api/catalog/menus/${selectedMenu.id}/bom`, {
      headers: { Authorization: `Bearer ${token}` },
    })
      .then((r) => r.json())
      .then((json) => setBom(Array.isArray(json) ? json : json.data ?? []))
      .catch(console.error)
      .finally(() => setBomLoading(false));
  }, [selectedMenu, session]);

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
  const categories = ["all", ...Array.from(new Set(menus.map((m) => m.category)))];
  const filtered = menus.filter((m) => {
    const matchSearch =
      m.name.toLowerCase().includes(search.toLowerCase()) ||
      m.category.toLowerCase().includes(search.toLowerCase());
    const matchCat = filterCategory === "all" || m.category === filterCategory;
    return matchSearch && matchCat;
  });

  // ─── Helpers ────────────────────────────────────────────────────
  function openCreate() {
    setEditingMenu(null);
    setFormName(""); setFormDescription(""); setFormPrice("");
    setFormCategory(""); setFormAvailable(true);
    setShowModal(true);
  }

  function openEdit(menu: MenuItem) {
    setEditingMenu(menu);
    setFormName(menu.name);
    setFormDescription(menu.description ?? "");
    setFormPrice(String(menu.price));
    setFormCategory(menu.category);
    setFormAvailable(menu.is_available);
    setShowModal(true);
  }

  async function handleSave(e: FormEvent) {
    e.preventDefault();
    if (!formName.trim() || !formPrice || saving) return;
    setSaving(true);

    const payload = {
      name: formName.trim(),
      description: formDescription.trim(),
      price: parseFloat(formPrice),
      category: formCategory.trim(),
      is_available: formAvailable,
    };

    try {
      let res: Response;
      if (editingMenu) {
        res = await fetch(`/api/catalog/menus/${editingMenu.id}`, {
          method: "PUT",
          headers: { "Content-Type": "application/json", Authorization: `Bearer ${token}` },
          body: JSON.stringify(payload),
        });
      } else {
        res = await fetch("/api/catalog/menus", {
          method: "POST",
          headers: { "Content-Type": "application/json", Authorization: `Bearer ${token}` },
          body: JSON.stringify(payload),
        });
      }

      if (!res.ok) {
        const err = await res.json().catch(() => null);
        alert((editingMenu ? "แก้ไข" : "สร้าง") + "เมนูไม่สำเร็จ: " + (err?.message ?? res.statusText));
        return;
      }

      const json = await res.json();
      const saved: MenuItem = json.data ?? json;
      if (editingMenu) {
        setMenus((prev) => prev.map((m) => (m.id === saved.id ? saved : m)));
        if (selectedMenu?.id === saved.id) setSelectedMenu(saved);
      } else {
        setMenus((prev) => [...prev, saved]);
      }
      setShowModal(false);
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
      const res = await fetch(`/api/catalog/menus/${deletingId}`, {
        method: "DELETE",
        headers: { Authorization: `Bearer ${token}` },
      });
      if (!res.ok) {
        const err = await res.json().catch(() => null);
        alert("ลบไม่สำเร็จ: " + (err?.message ?? res.statusText));
        return;
      }
      setMenus((prev) => prev.filter((m) => m.id !== deletingId));
      if (selectedMenu?.id === deletingId) setSelectedMenu(null);
    } catch (err) {
      console.error(err);
      alert("เกิดข้อผิดพลาด");
    } finally {
      setDeletingId(null);
    }
  }

  async function toggleAvailable(menu: MenuItem) {
    const updated = { ...menu, is_available: !menu.is_available };
    setMenus((prev) => prev.map((m) => (m.id === menu.id ? updated : m)));
    try {
      const res = await fetch(`/api/catalog/menus/${menu.id}`, {
        method: "PUT",
        headers: { "Content-Type": "application/json", Authorization: `Bearer ${token}` },
        body: JSON.stringify({ name: menu.name, description: menu.description, price: menu.price, category: menu.category, is_available: !menu.is_available }),
      });
      if (!res.ok) setMenus((prev) => prev.map((m) => (m.id === menu.id ? menu : m)));
    } catch {
      setMenus((prev) => prev.map((m) => (m.id === menu.id ? menu : m)));
    }
  }

  // ─── BOM handlers ───────────────────────────────────────────────
  async function handleAddBom(e: FormEvent) {
    e.preventDefault();
    if (!selectedMenu || !bomIngredientId || !bomQuantity || addingBom) return;
    setAddingBom(true);
    try {
      const res = await fetch(`/api/catalog/menus/${selectedMenu.id}/bom`, {
        method: "POST",
        headers: { "Content-Type": "application/json", Authorization: `Bearer ${token}` },
        body: JSON.stringify({ ingredient_id: Number(bomIngredientId), quantity: parseFloat(bomQuantity) }),
      });
      if (!res.ok) {
        const err = await res.json().catch(() => null);
        alert("เพิ่ม BOM ไม่สำเร็จ: " + (err?.message ?? res.statusText));
        return;
      }
      const json = await res.json();
      const newItem: BOMItem = json.data ?? json;
      setBom((prev) => [...prev, newItem]);
      setBomIngredientId(""); setBomQuantity(""); setShowAddBom(false);
    } catch (err) {
      console.error(err);
      alert("เกิดข้อผิดพลาด");
    } finally {
      setAddingBom(false);
    }
  }

  async function handleDeleteBom(bomId: number) {
    if (!selectedMenu) return;
    try {
      const res = await fetch(`/api/catalog/menus/${selectedMenu.id}/bom/${bomId}`, {
        method: "DELETE",
        headers: { Authorization: `Bearer ${token}` },
      });
      if (!res.ok) {
        const err = await res.json().catch(() => null);
        alert("ลบ BOM ไม่สำเร็จ: " + (err?.message ?? res.statusText));
        return;
      }
      setBom((prev) => prev.filter((b) => b.id !== bomId));
    } catch (err) {
      console.error(err);
    }
  }

  function ingredientName(id?: number) {
    if (!id) return "—";
    return ingredients.find((i) => i.id === id)?.name ?? `#${id}`;
  }

  function ingredientUnit(id?: number) {
    if (!id) return "";
    return ingredients.find((i) => i.id === id)?.unit ?? "";
  }

  // ─── Render ─────────────────────────────────────────────────────
  return (
    <div className="max-w-6xl mx-auto px-4 py-8">
      {/* Header */}
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold">จัดการเมนูอาหาร</h1>
          <p className="text-sm text-gray-500 mt-1">เพิ่ม แก้ไข ลบ และจัดการสูตรอาหาร (BOM)</p>
        </div>
        <button
          onClick={openCreate}
          className="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 transition text-sm font-medium"
        >
          + เพิ่มเมนู
        </button>
      </div>

      {/* Filters */}
      <div className="flex flex-col sm:flex-row gap-3 mb-6">
        <input
          type="text"
          placeholder="ค้นหาเมนู..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="flex-1 border rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
        />
        <select
          value={filterCategory}
          onChange={(e) => setFilterCategory(e.target.value)}
          className="border rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
        >
          {categories.map((c) => (
            <option key={c} value={c}>{c === "all" ? "ทุกหมวด" : c}</option>
          ))}
        </select>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-3 gap-4 mb-6">
        {[
          { label: "เมนูทั้งหมด", value: menus.length, color: "blue" },
          { label: "เปิดขาย", value: menus.filter((m) => m.is_available).length, color: "green" },
          { label: "ปิดขาย", value: menus.filter((m) => !m.is_available).length, color: "gray" },
        ].map(({ label, value, color }) => (
          <div key={label} className="bg-white rounded-lg shadow p-4 text-center">
            <p className={`text-2xl font-bold text-${color}-600`}>{value}</p>
            <p className="text-xs text-gray-500 mt-1">{label}</p>
          </div>
        ))}
      </div>

      {/* Main layout */}
      <div className="flex flex-col lg:flex-row gap-6">
        {/* Menu list */}
        <div className={selectedMenu ? "lg:w-1/2" : "w-full"}>
          {filtered.length === 0 ? (
            <div className="bg-white rounded-lg shadow p-8 text-center text-gray-400">ไม่พบเมนู</div>
          ) : (
            <div className="space-y-3">
              {filtered.map((menu) => (
                <div
                  key={menu.id}
                  className={`bg-white rounded-lg shadow p-4 border-l-4 ${
                    menu.is_available ? "border-green-500" : "border-gray-300"
                  } ${selectedMenu?.id === menu.id ? "ring-2 ring-blue-400" : ""}`}
                >
                  <div className="flex items-start justify-between">
                    <div
                      className="flex-1 cursor-pointer"
                      onClick={() => setSelectedMenu(selectedMenu?.id === menu.id ? null : menu)}
                    >
                      <div className="flex items-center gap-2 flex-wrap">
                        <h3 className="font-semibold text-gray-800">{menu.name}</h3>
                        <span className={`text-xs px-2 py-0.5 rounded-full font-medium ${
                          menu.is_available ? "bg-green-100 text-green-700" : "bg-gray-100 text-gray-500"
                        }`}>
                          {menu.is_available ? "เปิดขาย" : "ปิดขาย"}
                        </span>
                        <span className="text-xs bg-blue-50 text-blue-700 px-2 py-0.5 rounded">
                          {menu.category}
                        </span>
                      </div>
                      {menu.description && (
                        <p className="text-sm text-gray-500 mt-1">{menu.description}</p>
                      )}
                      <p className="text-sm font-semibold text-orange-600 mt-1">
                        ฿{menu.price.toFixed(2)}
                      </p>
                    </div>
                    <div className="flex items-center gap-1 ml-3 shrink-0">
                      <button
                        onClick={() => toggleAvailable(menu)}
                        className={`p-1.5 rounded text-xs font-medium transition ${
                          menu.is_available
                            ? "text-yellow-600 hover:bg-yellow-50"
                            : "text-green-600 hover:bg-green-50"
                        }`}
                      >
                        {menu.is_available ? "ปิด" : "เปิด"}
                      </button>
                      <button
                        onClick={() => openEdit(menu)}
                        className="p-1.5 rounded text-blue-600 hover:bg-blue-50 text-xs font-medium transition"
                      >
                        แก้ไข
                      </button>
                      <button
                        onClick={() => setDeletingId(menu.id)}
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

        {/* BOM detail panel */}
        {selectedMenu && (
          <div className="lg:w-1/2">
            <div className="bg-white rounded-lg shadow p-6 sticky top-4">
              <div className="flex items-center justify-between mb-4">
                <div>
                  <h2 className="text-lg font-bold text-gray-800">{selectedMenu.name}</h2>
                  <p className="text-xs text-gray-400 mt-0.5">สูตรอาหาร (BOM)</p>
                </div>
                <button
                  onClick={() => setSelectedMenu(null)}
                  className="text-gray-400 hover:text-gray-600 text-lg"
                >
                  ✕
                </button>
              </div>

              {bomLoading ? (
                <p className="text-sm text-gray-400 py-4 text-center">กำลังโหลด...</p>
              ) : (
                <>
                  {bom.length === 0 ? (
                    <p className="text-sm text-gray-400 mb-4">ยังไม่มีสูตรอาหาร</p>
                  ) : (
                    <div className="space-y-2 mb-4">
                      {bom.map((item) => (
                        <div
                          key={item.id}
                          className="flex items-center justify-between bg-gray-50 rounded px-3 py-2"
                        >
                          <div>
                            <p className="text-sm font-medium text-gray-700">
                              {item.ingredient?.name ?? ingredientName(item.ingredient_id)}
                              {item.sub_menu_item && (
                                <span className="text-blue-600"> [{item.sub_menu_item.name}]</span>
                              )}
                            </p>
                            <p className="text-xs text-gray-400">
                              {item.quantity} {item.ingredient?.unit ?? ingredientUnit(item.ingredient_id)}
                            </p>
                          </div>
                          <button
                            onClick={() => handleDeleteBom(item.id)}
                            className="text-xs text-red-500 hover:text-red-700"
                          >
                            ลบ
                          </button>
                        </div>
                      ))}
                    </div>
                  )}

                  {/* Add BOM */}
                  {showAddBom ? (
                    <form onSubmit={handleAddBom} className="border-t pt-4 space-y-3">
                      <p className="text-sm font-medium text-gray-600">เพิ่มวัตถุดิบ</p>
                      <select
                        value={bomIngredientId}
                        onChange={(e) => setBomIngredientId(e.target.value)}
                        className="w-full border rounded px-2 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                        required
                      >
                        <option value="">เลือกวัตถุดิบ...</option>
                        {ingredients.map((ing) => (
                          <option key={ing.id} value={ing.id}>
                            {ing.name} ({ing.unit})
                          </option>
                        ))}
                      </select>
                      <div className="flex gap-2">
                        <input
                          type="number"
                          step="0.001"
                          min="0.001"
                          placeholder="ปริมาณ"
                          value={bomQuantity}
                          onChange={(e) => setBomQuantity(e.target.value)}
                          className="flex-1 border rounded px-2 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                          required
                        />
                        <button
                          type="submit"
                          disabled={addingBom}
                          className="bg-blue-600 text-white px-3 py-1.5 rounded text-sm hover:bg-blue-700 disabled:opacity-50 transition"
                        >
                          เพิ่ม
                        </button>
                        <button
                          type="button"
                          onClick={() => { setShowAddBom(false); setBomIngredientId(""); setBomQuantity(""); }}
                          className="px-3 py-1.5 rounded text-sm text-gray-600 hover:bg-gray-100 transition"
                        >
                          ยกเลิก
                        </button>
                      </div>
                    </form>
                  ) : (
                    <button
                      onClick={() => setShowAddBom(true)}
                      className="w-full mt-2 border-2 border-dashed border-gray-200 rounded-lg py-2 text-sm text-gray-500 hover:border-blue-300 hover:text-blue-600 transition"
                    >
                      + เพิ่มวัตถุดิบ
                    </button>
                  )}
                </>
              )}
            </div>
          </div>
        )}
      </div>

      {/* ─── Create/Edit Modal ─────────────────────────────────── */}
      {showModal && (
        <div className="fixed inset-0 bg-black/40 flex items-center justify-center z-50">
          <div className="bg-white rounded-xl shadow-lg w-full max-w-md mx-4 p-6">
            <h2 className="text-lg font-bold mb-4">
              {editingMenu ? "แก้ไขเมนู" : "เพิ่มเมนูใหม่"}
            </h2>
            <form onSubmit={handleSave} className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">ชื่อเมนู *</label>
                <input
                  type="text"
                  value={formName}
                  onChange={(e) => setFormName(e.target.value)}
                  className="w-full border rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="เช่น ข้าวผัดกระเพราหมูสับ"
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
                  placeholder="คำอธิบายเมนู"
                />
              </div>
              <div className="flex gap-3">
                <div className="flex-1">
                  <label className="block text-sm font-medium text-gray-700 mb-1">ราคา (฿) *</label>
                  <input
                    type="number"
                    step="0.01"
                    min="0"
                    value={formPrice}
                    onChange={(e) => setFormPrice(e.target.value)}
                    className="w-full border rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                    placeholder="0.00"
                    required
                  />
                </div>
                <div className="flex-1">
                  <label className="block text-sm font-medium text-gray-700 mb-1">หมวดหมู่ *</label>
                  <input
                    type="text"
                    value={formCategory}
                    onChange={(e) => setFormCategory(e.target.value)}
                    className="w-full border rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                    placeholder="เช่น อาหารจานเดียว"
                    required
                  />
                </div>
              </div>
              <div className="flex items-center gap-2">
                <input
                  type="checkbox"
                  id="is_available"
                  checked={formAvailable}
                  onChange={(e) => setFormAvailable(e.target.checked)}
                  className="rounded"
                />
                <label htmlFor="is_available" className="text-sm text-gray-700">เปิดขาย</label>
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
                  {editingMenu ? "บันทึก" : "เพิ่มเมนู"}
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
            <p className="text-sm text-gray-500 mb-4">
              คุณต้องการลบเมนู &ldquo;{menus.find((m) => m.id === deletingId)?.name}&rdquo; ใช่หรือไม่?
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
                ลบเมนู
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
