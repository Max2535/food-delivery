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

interface MenuAddOn {
  id: number;
  menu_item_id: number;
  ingredient_id: number;
  quantity: number;
  extra_price: number;
  is_available: boolean;
  ingredient?: Ingredient;
}

interface BOMChoiceOption {
  id: number;
  group_id: number;
  ingredient_id: number;
  quantity: number;
  extra_price: number;
  ingredient?: Ingredient;
}

interface BOMChoiceGroup {
  id: number;
  menu_item_id: number;
  name: string;
  is_required: boolean;
  min_choices: number;
  max_choices: number;
  options?: BOMChoiceOption[];
}

interface MenuPortionSize {
  id: number;
  menu_item_id: number;
  name: string;
  quantity_multiplier: number;
  extra_price: number;
  is_default: boolean;
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

type DetailTab = "bom" | "addons" | "choices" | "portions";

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

  // Detail panel
  const [selectedMenu, setSelectedMenu] = useState<MenuItem | null>(null);
  const [activeTab, setActiveTab] = useState<DetailTab>("bom");

  // BOM state
  const [bom, setBom] = useState<BOMItem[]>([]);
  const [bomLoading, setBomLoading] = useState(false);
  const [showAddBom, setShowAddBom] = useState(false);
  const [bomType, setBomType] = useState<"ingredient" | "sub_menu">("ingredient");
  const [bomIngredientId, setBomIngredientId] = useState("");
  const [bomSubMenuId, setBomSubMenuId] = useState("");
  const [bomQuantity, setBomQuantity] = useState("");
  const [addingBom, setAddingBom] = useState(false);

  // Add-on state
  const [addons, setAddons] = useState<MenuAddOn[]>([]);
  const [addonsLoading, setAddonsLoading] = useState(false);
  const [showAddAddon, setShowAddAddon] = useState(false);
  const [addonIngredientId, setAddonIngredientId] = useState("");
  const [addonQuantity, setAddonQuantity] = useState("");
  const [addonExtraPrice, setAddonExtraPrice] = useState("");
  const [addingAddon, setAddingAddon] = useState(false);

  // Choice state
  const [choices, setChoices] = useState<BOMChoiceGroup[]>([]);
  const [choicesLoading, setChoicesLoading] = useState(false);
  const [showAddGroup, setShowAddGroup] = useState(false);
  const [groupName, setGroupName] = useState("");
  const [groupRequired, setGroupRequired] = useState(true);
  const [groupMin, setGroupMin] = useState("1");
  const [groupMax, setGroupMax] = useState("1");
  const [addingGroup, setAddingGroup] = useState(false);
  // Add option to existing group
  const [addingOptionToGroup, setAddingOptionToGroup] = useState<number | null>(null);
  const [optionIngredientId, setOptionIngredientId] = useState("");
  const [optionQuantity, setOptionQuantity] = useState("");
  const [optionExtraPrice, setOptionExtraPrice] = useState("");
  const [addingOption, setAddingOption] = useState(false);

  // Portion state
  const [portions, setPortions] = useState<MenuPortionSize[]>([]);
  const [portionsLoading, setPortionsLoading] = useState(false);
  const [showAddPortion, setShowAddPortion] = useState(false);
  const [portionName, setPortionName] = useState("");
  const [portionMultiplier, setPortionMultiplier] = useState("1.0");
  const [portionExtraPrice, setPortionExtraPrice] = useState("0");
  const [portionIsDefault, setPortionIsDefault] = useState(false);
  const [addingPortion, setAddingPortion] = useState(false);

  // Delete confirmation
  const [deletingId, setDeletingId] = useState<number | null>(null);

  useEffect(() => {
    if (status === "unauthenticated") router.replace("/auth/login");
  }, [status, router]);

  useEffect(() => {
    const token = (session as any)?.accessToken;
    if (status !== "authenticated" || !token) return;
    const controller = new AbortController();
    Promise.all([
      fetch("/api/catalog/menus", { headers: { Authorization: `Bearer ${token}` }, signal: controller.signal }),
      fetch("/api/catalog/ingredients", { headers: { Authorization: `Bearer ${token}` }, signal: controller.signal }),
    ])
      .then(async ([menusRes, ingredientsRes]) => {
        if (menusRes.ok) {
          const j = await menusRes.json();
          setMenus(Array.isArray(j) ? j : j.menu_items ?? j.data ?? []);
        }
        if (ingredientsRes.ok) {
          const j = await ingredientsRes.json();
          setIngredients(Array.isArray(j) ? j : j.ingredients ?? j.data ?? []);
        }
      })
      .catch((err) => { if (err.name !== "AbortError") console.error(err); });
    return () => controller.abort();
  }, [status, session]);

  // Load detail data when menu or tab changes
  useEffect(() => {
    if (!selectedMenu) { setBom([]); setAddons([]); setChoices([]); setPortions([]); return; }
    const token = (session as any)?.accessToken;
    if (!token) return;

    if (activeTab === "bom") {
      setBomLoading(true);
      fetch(`/api/catalog/menus/${selectedMenu.id}/bom`, { headers: { Authorization: `Bearer ${token}` } })
        .then((r) => r.json())
        .then((j) => setBom(Array.isArray(j) ? j : j.bom_items ?? []))
        .catch(console.error)
        .finally(() => setBomLoading(false));
    }
    if (activeTab === "addons") {
      setAddonsLoading(true);
      fetch(`/api/catalog/menus/${selectedMenu.id}/addons`, { headers: { Authorization: `Bearer ${token}` } })
        .then((r) => r.json())
        .then((j) => setAddons(Array.isArray(j) ? j : j.addons ?? []))
        .catch(console.error)
        .finally(() => setAddonsLoading(false));
    }
    if (activeTab === "choices") {
      setChoicesLoading(true);
      fetch(`/api/catalog/menus/${selectedMenu.id}/choices`, { headers: { Authorization: `Bearer ${token}` } })
        .then((r) => r.json())
        .then((j) => setChoices(Array.isArray(j) ? j : j.choice_groups ?? []))
        .catch(console.error)
        .finally(() => setChoicesLoading(false));
    }
    if (activeTab === "portions") {
      setPortionsLoading(true);
      fetch(`/api/catalog/menus/${selectedMenu.id}/portions`, { headers: { Authorization: `Bearer ${token}` } })
        .then((r) => r.json())
        .then((j) => setPortions(Array.isArray(j) ? j : j.portions ?? []))
        .catch(console.error)
        .finally(() => setPortionsLoading(false));
    }
  }, [selectedMenu, activeTab, session]);

  if (status === "loading") {
    return (
      <div className="min-h-[80vh] flex items-center justify-center">
        <p className="text-gray-500">กำลังโหลด...</p>
      </div>
    );
  }
  if (!session) return null;

  const token = (session as any)?.accessToken;

  // ─── Derived data ─────────────────────────────────────────────
  const categories = ["all", ...Array.from(new Set(menus.map((m) => m.category)))];
  const filtered = menus.filter((m) => {
    const matchSearch =
      m.name.toLowerCase().includes(search.toLowerCase()) ||
      m.category.toLowerCase().includes(search.toLowerCase());
    const matchCat = filterCategory === "all" || m.category === filterCategory;
    return matchSearch && matchCat;
  });

  function ingredientName(id?: number) {
    if (!id) return "—";
    return ingredients.find((i) => i.id === id)?.name ?? `#${id}`;
  }
  function ingredientUnit(id?: number) {
    if (!id) return "";
    return ingredients.find((i) => i.id === id)?.unit ?? "";
  }

  // ─── Menu CRUD ────────────────────────────────────────────────
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
    const payload = { name: formName.trim(), description: formDescription.trim(), price: parseFloat(formPrice), category: formCategory.trim(), is_available: formAvailable };
    try {
      let res: Response;
      if (editingMenu) {
        res = await fetch(`/api/catalog/menus/${editingMenu.id}`, { method: "PUT", headers: { "Content-Type": "application/json", Authorization: `Bearer ${token}` }, body: JSON.stringify(payload) });
      } else {
        res = await fetch("/api/catalog/menus", { method: "POST", headers: { "Content-Type": "application/json", Authorization: `Bearer ${token}` }, body: JSON.stringify(payload) });
      }
      if (!res.ok) { const err = await res.json().catch(() => null); alert((editingMenu ? "แก้ไข" : "สร้าง") + "เมนูไม่สำเร็จ: " + (err?.message ?? res.statusText)); return; }
      const json = await res.json();
      const saved: MenuItem = json.data ?? json;
      if (editingMenu) {
        setMenus((prev) => prev.map((m) => (m.id === saved.id ? saved : m)));
        if (selectedMenu?.id === saved.id) setSelectedMenu(saved);
      } else {
        setMenus((prev) => [...prev, saved]);
        setSelectedMenu(saved);
        setActiveTab("bom");
      }
      setShowModal(false);
    } catch (err) { console.error(err); alert("เกิดข้อผิดพลาด"); }
    finally { setSaving(false); }
  }

  async function handleDelete() {
    if (!deletingId) return;
    try {
      const res = await fetch(`/api/catalog/menus/${deletingId}`, { method: "DELETE", headers: { Authorization: `Bearer ${token}` } });
      if (!res.ok) { const err = await res.json().catch(() => null); alert("ลบไม่สำเร็จ: " + (err?.message ?? res.statusText)); return; }
      setMenus((prev) => prev.filter((m) => m.id !== deletingId));
      if (selectedMenu?.id === deletingId) setSelectedMenu(null);
    } catch (err) { console.error(err); alert("เกิดข้อผิดพลาด"); }
    finally { setDeletingId(null); }
  }

  async function toggleAvailable(menu: MenuItem) {
    const updated = { ...menu, is_available: !menu.is_available };
    setMenus((prev) => prev.map((m) => (m.id === menu.id ? updated : m)));
    try {
      const res = await fetch(`/api/catalog/menus/${menu.id}`, { method: "PUT", headers: { "Content-Type": "application/json", Authorization: `Bearer ${token}` }, body: JSON.stringify({ name: menu.name, description: menu.description, price: menu.price, category: menu.category, is_available: !menu.is_available }) });
      if (!res.ok) setMenus((prev) => prev.map((m) => (m.id === menu.id ? menu : m)));
    } catch { setMenus((prev: MenuItem[]) => prev.map((m: MenuItem) => (m.id === menu.id ? menu : m))); }
  }

  function selectMenu(menu: MenuItem) {
    if (selectedMenu?.id === menu.id) { setSelectedMenu(null); return; }
    setSelectedMenu(menu);
    setActiveTab("bom");
    setShowAddBom(false); setShowAddAddon(false); setShowAddGroup(false); setShowAddPortion(false);
    setAddingOptionToGroup(null); setBomType("ingredient"); setBomSubMenuId("");
  }

  // ─── BOM handlers ────────────────────────────────────────────
  async function handleAddBom(e: FormEvent) {
    e.preventDefault();
    if (!selectedMenu || !bomQuantity || addingBom) return;
    if (bomType === "ingredient" && !bomIngredientId) return;
    if (bomType === "sub_menu" && !bomSubMenuId) return;
    setAddingBom(true);
    try {
      const payload: Record<string, unknown> = { quantity: parseFloat(bomQuantity) };
      if (bomType === "ingredient") {
        payload.ingredient_id = Number(bomIngredientId);
      } else {
        payload.sub_menu_item_id = Number(bomSubMenuId);
      }
      const res = await fetch(`/api/catalog/menus/${selectedMenu.id}/bom`, { method: "POST", headers: { "Content-Type": "application/json", Authorization: `Bearer ${token}` }, body: JSON.stringify(payload) });
      if (!res.ok) { const err = await res.json().catch(() => null); alert("เพิ่ม BOM ไม่สำเร็จ: " + (err?.message ?? res.statusText)); return; }
      const json = await res.json();
      setBom((prev) => [...prev, json.data ?? json]);
      setBomIngredientId(""); setBomSubMenuId(""); setBomQuantity(""); setShowAddBom(false);
    } catch (err) { console.error(err); alert("เกิดข้อผิดพลาด"); }
    finally { setAddingBom(false); }
  }

  async function handleDeleteBom(bomId: number) {
    if (!selectedMenu) return;
    try {
      const res = await fetch(`/api/catalog/menus/${selectedMenu.id}/bom/${bomId}`, { method: "DELETE", headers: { Authorization: `Bearer ${token}` } });
      if (!res.ok) { const err = await res.json().catch(() => null); alert("ลบ BOM ไม่สำเร็จ: " + (err?.message ?? res.statusText)); return; }
      setBom((prev) => prev.filter((b) => b.id !== bomId));
    } catch (err) { console.error(err); }
  }

  // ─── Add-on handlers ─────────────────────────────────────────
  async function handleAddAddon(e: FormEvent) {
    e.preventDefault();
    if (!selectedMenu || !addonIngredientId || !addonQuantity || addingAddon) return;
    setAddingAddon(true);
    try {
      const res = await fetch(`/api/catalog/menus/${selectedMenu.id}/addons`, { method: "POST", headers: { "Content-Type": "application/json", Authorization: `Bearer ${token}` }, body: JSON.stringify({ ingredient_id: Number(addonIngredientId), quantity: parseFloat(addonQuantity), extra_price: parseFloat(addonExtraPrice || "0") }) });
      if (!res.ok) { const err = await res.json().catch(() => null); alert("เพิ่ม Topping ไม่สำเร็จ: " + (err?.message ?? res.statusText)); return; }
      const json = await res.json();
      setAddons((prev: MenuAddOn[]) => [...prev, json.data ?? json]);
      setAddonIngredientId(""); setAddonQuantity(""); setAddonExtraPrice(""); setShowAddAddon(false);
    } catch (err) { console.error(err); alert("เกิดข้อผิดพลาด"); }
    finally { setAddingAddon(false); }
  }

  async function handleDeleteAddon(addonId: number) {
    if (!selectedMenu) return;
    try {
      const res = await fetch(`/api/catalog/menus/${selectedMenu.id}/addons/${addonId}`, { method: "DELETE", headers: { Authorization: `Bearer ${token}` } });
      if (!res.ok) { const err = await res.json().catch(() => null); alert("ลบ Topping ไม่สำเร็จ: " + (err?.message ?? res.statusText)); return; }
      setAddons((prev: MenuAddOn[]) => prev.filter((a: MenuAddOn) => a.id !== addonId));
    } catch (err) { console.error(err); }
  }

  // ─── Choice handlers ─────────────────────────────────────────
  async function handleAddGroup(e: FormEvent) {
    e.preventDefault();
    if (!selectedMenu || !groupName.trim() || addingGroup) return;
    setAddingGroup(true);
    try {
      const res = await fetch(`/api/catalog/menus/${selectedMenu.id}/choices`, { method: "POST", headers: { "Content-Type": "application/json", Authorization: `Bearer ${token}` }, body: JSON.stringify({ name: groupName.trim(), is_required: groupRequired, min_choices: Number(groupMin), max_choices: Number(groupMax) }) });
      if (!res.ok) { const err = await res.json().catch(() => null); alert("เพิ่มกลุ่มไม่สำเร็จ: " + (err?.message ?? res.statusText)); return; }
      const json = await res.json();
      setChoices((prev) => [...prev, { ...(json.data ?? json), options: [] }]);
      setGroupName(""); setGroupRequired(true); setGroupMin("1"); setGroupMax("1"); setShowAddGroup(false);
    } catch (err) { console.error(err); alert("เกิดข้อผิดพลาด"); }
    finally { setAddingGroup(false); }
  }

  async function handleDeleteGroup(groupId: number) {
    if (!selectedMenu) return;
    try {
      const res = await fetch(`/api/catalog/menus/${selectedMenu.id}/choices/${groupId}`, { method: "DELETE", headers: { Authorization: `Bearer ${token}` } });
      if (!res.ok) { const err = await res.json().catch(() => null); alert("ลบกลุ่มไม่สำเร็จ: " + (err?.message ?? res.statusText)); return; }
      setChoices((prev: BOMChoiceGroup[]) => prev.filter((g: BOMChoiceGroup) => g.id !== groupId));
    } catch (err) { console.error(err); }
  }

  async function handleAddOption(e: FormEvent) {
    e.preventDefault();
    if (!selectedMenu || addingOptionToGroup === null || !optionIngredientId || !optionQuantity || addingOption) return;
    setAddingOption(true);
    try {
      const res = await fetch(`/api/catalog/menus/${selectedMenu.id}/choices/${addingOptionToGroup}/options`, { method: "POST", headers: { "Content-Type": "application/json", Authorization: `Bearer ${token}` }, body: JSON.stringify({ ingredient_id: Number(optionIngredientId), quantity: parseFloat(optionQuantity), extra_price: parseFloat(optionExtraPrice || "0") }) });
      if (!res.ok) { const err = await res.json().catch(() => null); alert("เพิ่มตัวเลือกไม่สำเร็จ: " + (err?.message ?? res.statusText)); return; }
      const json = await res.json();
      const newOption: BOMChoiceOption = json.data ?? json;
      setChoices((prev: BOMChoiceGroup[]) => prev.map((g: BOMChoiceGroup) => g.id === addingOptionToGroup ? { ...g, options: [...(g.options ?? []), newOption] } : g));
      setOptionIngredientId(""); setOptionQuantity(""); setOptionExtraPrice(""); setAddingOptionToGroup(null);
    } catch (err) { console.error(err); alert("เกิดข้อผิดพลาด"); }
    finally { setAddingOption(false); }
  }

  async function handleDeleteOption(groupId: number, optionId: number) {
    if (!selectedMenu) return;
    try {
      const res = await fetch(`/api/catalog/menus/${selectedMenu.id}/choices/${groupId}/options/${optionId}`, { method: "DELETE", headers: { Authorization: `Bearer ${token}` } });
      if (!res.ok) { const err = await res.json().catch(() => null); alert("ลบตัวเลือกไม่สำเร็จ: " + (err?.message ?? res.statusText)); return; }
      setChoices((prev: BOMChoiceGroup[]) => prev.map((g: BOMChoiceGroup) => g.id === groupId ? { ...g, options: (g.options ?? []).filter((o: BOMChoiceOption) => o.id !== optionId) } : g));
    } catch (err) { console.error(err); }
  }

  // ─── Portion handlers ─────────────────────────────────────────
  async function handleAddPortion(e: FormEvent) {
    e.preventDefault();
    if (!selectedMenu || !portionName.trim() || addingPortion) return;
    setAddingPortion(true);
    try {
      const res = await fetch(`/api/catalog/menus/${selectedMenu.id}/portions`, { method: "POST", headers: { "Content-Type": "application/json", Authorization: `Bearer ${token}` }, body: JSON.stringify({ name: portionName.trim(), quantity_multiplier: parseFloat(portionMultiplier || "1"), extra_price: parseFloat(portionExtraPrice || "0"), is_default: portionIsDefault }) });
      if (!res.ok) { const err = await res.json().catch(() => null); alert("เพิ่มขนาดไม่สำเร็จ: " + (err?.message ?? res.statusText)); return; }
      const json = await res.json();
      setPortions((prev: MenuPortionSize[]) => [...prev, json.data ?? json]);
      setPortionName(""); setPortionMultiplier("1.0"); setPortionExtraPrice("0"); setPortionIsDefault(false); setShowAddPortion(false);
    } catch (err) { console.error(err); alert("เกิดข้อผิดพลาด"); }
    finally { setAddingPortion(false); }
  }

  async function handleDeletePortion(portionId: number) {
    if (!selectedMenu) return;
    try {
      const res = await fetch(`/api/catalog/menus/${selectedMenu.id}/portions/${portionId}`, { method: "DELETE", headers: { Authorization: `Bearer ${token}` } });
      if (!res.ok) { const err = await res.json().catch(() => null); alert("ลบขนาดไม่สำเร็จ: " + (err?.message ?? res.statusText)); return; }
      setPortions((prev: MenuPortionSize[]) => prev.filter((p: MenuPortionSize) => p.id !== portionId));
    } catch (err) { console.error(err); }
  }

  // ─── Render ──────────────────────────────────────────────────
  return (
    <div className="max-w-6xl mx-auto px-4 py-8">
      {/* Header */}
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold">จัดการเมนูอาหาร</h1>
          <p className="text-sm text-gray-500 mt-1">เพิ่ม แก้ไข และจัดการ BOM / Topping / ตัวเลือก / ขนาด</p>
        </div>
        <button onClick={openCreate} className="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 transition text-sm font-medium">
          + เพิ่มเมนู
        </button>
      </div>

      {/* Filters */}
      <div className="flex flex-col sm:flex-row gap-3 mb-6">
        <input type="text" placeholder="ค้นหาเมนู..." value={search} onChange={(e) => setSearch(e.target.value)}
          className="flex-1 border rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
        <select value={filterCategory} onChange={(e) => setFilterCategory(e.target.value)}
          className="border rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500">
          {categories.map((c) => <option key={c} value={c}>{c === "all" ? "ทุกหมวด" : c}</option>)}
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
                <div key={menu.id}
                  className={`bg-white rounded-lg shadow p-4 border-l-4 ${menu.is_available ? "border-green-500" : "border-gray-300"} ${selectedMenu?.id === menu.id ? "ring-2 ring-blue-400" : ""}`}>
                  <div className="flex items-start justify-between">
                    <div className="flex-1 cursor-pointer" onClick={() => selectMenu(menu)}>
                      <div className="flex items-center gap-2 flex-wrap">
                        <h3 className="font-semibold text-gray-800">{menu.name}</h3>
                        <span className={`text-xs px-2 py-0.5 rounded-full font-medium ${menu.is_available ? "bg-green-100 text-green-700" : "bg-gray-100 text-gray-500"}`}>
                          {menu.is_available ? "เปิดขาย" : "ปิดขาย"}
                        </span>
                        <span className="text-xs bg-blue-50 text-blue-700 px-2 py-0.5 rounded">{menu.category}</span>
                      </div>
                      {menu.description && <p className="text-sm text-gray-500 mt-1">{menu.description}</p>}
                      <p className="text-sm font-semibold text-orange-600 mt-1">฿{menu.price.toFixed(2)}</p>
                    </div>
                    <div className="flex items-center gap-1 ml-3 shrink-0">
                      <button onClick={() => toggleAvailable(menu)}
                        className={`p-1.5 rounded text-xs font-medium transition ${menu.is_available ? "text-yellow-600 hover:bg-yellow-50" : "text-green-600 hover:bg-green-50"}`}>
                        {menu.is_available ? "ปิด" : "เปิด"}
                      </button>
                      <button onClick={() => openEdit(menu)} className="p-1.5 rounded text-blue-600 hover:bg-blue-50 text-xs font-medium transition">แก้ไข</button>
                      <button onClick={() => setDeletingId(menu.id)} className="p-1.5 rounded text-red-600 hover:bg-red-50 text-xs font-medium transition">ลบ</button>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Detail panel */}
        {selectedMenu && (
          <div className="lg:w-1/2">
            <div className="bg-white rounded-lg shadow sticky top-4">
              {/* Panel header */}
              <div className="flex items-center justify-between px-6 pt-5 pb-3 border-b">
                <div>
                  <h2 className="text-base font-bold text-gray-800">{selectedMenu.name}</h2>
                  <p className="text-xs text-orange-600 mt-0.5">฿{selectedMenu.price.toFixed(2)}</p>
                </div>
                <button onClick={() => setSelectedMenu(null)} className="text-gray-400 hover:text-gray-600 text-lg">✕</button>
              </div>

              {/* Tabs */}
              <div className="flex border-b text-sm">
                {(["bom", "addons", "choices", "portions"] as DetailTab[]).map((tab) => {
                  const labels: Record<DetailTab, string> = { bom: "BOM", addons: "Topping", choices: "ตัวเลือก", portions: "ขนาด" };
                  return (
                    <button key={tab} onClick={() => setActiveTab(tab)}
                      className={`flex-1 py-2.5 font-medium transition border-b-2 ${activeTab === tab ? "border-blue-600 text-blue-600" : "border-transparent text-gray-500 hover:text-gray-700"}`}>
                      {labels[tab]}
                    </button>
                  );
                })}
              </div>

              <div className="p-5">
                {/* ── BOM Tab ── */}
                {activeTab === "bom" && (
                  bomLoading ? <p className="text-sm text-gray-400 text-center py-6">กำลังโหลด...</p> : (
                    <>
                      {bom.length === 0 ? <p className="text-sm text-gray-400 mb-3">ยังไม่มีสูตรอาหาร</p> : (
                        <div className="space-y-2 mb-4">
                          {bom.map((item) => (
                            <div key={item.id} className="flex items-center justify-between bg-gray-50 rounded px-3 py-2">
                              <div>
                                {item.sub_menu_item_id ? (
                                  <>
                                    <div className="flex items-center gap-1.5">
                                      <span className="text-xs bg-blue-100 text-blue-700 px-1.5 py-0.5 rounded font-medium">เมนูย่อย</span>
                                      <p className="text-sm font-medium text-blue-700">
                                        {item.sub_menu_item?.name ?? menus.find((m) => m.id === item.sub_menu_item_id)?.name ?? `เมนู #${item.sub_menu_item_id}`}
                                      </p>
                                    </div>
                                    <p className="text-xs text-gray-400">x{item.quantity}</p>
                                  </>
                                ) : (
                                  <>
                                    <p className="text-sm font-medium text-gray-700">
                                      {item.ingredient?.name ?? ingredientName(item.ingredient_id)}
                                    </p>
                                    <p className="text-xs text-gray-400">{item.quantity} {item.ingredient?.unit ?? ingredientUnit(item.ingredient_id)}</p>
                                  </>
                                )}
                              </div>
                              <button onClick={() => handleDeleteBom(item.id)} className="text-xs text-red-500 hover:text-red-700">ลบ</button>
                            </div>
                          ))}
                        </div>
                      )}
                      {showAddBom ? (
                        <form onSubmit={handleAddBom} className="border-t pt-4 space-y-3">
                          <p className="text-sm font-medium text-gray-600">เพิ่มส่วนประกอบ BOM</p>
                          <div className="flex gap-2">
                            <button type="button" onClick={() => { setBomType("ingredient"); setBomSubMenuId(""); }}
                              className={`flex-1 py-1.5 rounded text-sm font-medium transition ${bomType === "ingredient" ? "bg-blue-600 text-white" : "bg-gray-100 text-gray-600 hover:bg-gray-200"}`}>
                              วัตถุดิบ
                            </button>
                            <button type="button" onClick={() => { setBomType("sub_menu"); setBomIngredientId(""); }}
                              className={`flex-1 py-1.5 rounded text-sm font-medium transition ${bomType === "sub_menu" ? "bg-blue-600 text-white" : "bg-gray-100 text-gray-600 hover:bg-gray-200"}`}>
                              เมนูย่อย (Sub-recipe)
                            </button>
                          </div>
                          {bomType === "ingredient" ? (
                            <select value={bomIngredientId} onChange={(e) => setBomIngredientId(e.target.value)}
                              className="w-full border rounded px-2 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" required>
                              <option value="">เลือกวัตถุดิบ...</option>
                              {ingredients.map((ing) => <option key={ing.id} value={ing.id}>{ing.name} ({ing.unit})</option>)}
                            </select>
                          ) : (
                            <select value={bomSubMenuId} onChange={(e) => setBomSubMenuId(e.target.value)}
                              className="w-full border rounded px-2 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" required>
                              <option value="">เลือกเมนูย่อย...</option>
                              {menus.filter((m) => m.id !== selectedMenu?.id).map((m) => <option key={m.id} value={m.id}>{m.name}</option>)}
                            </select>
                          )}
                          <div className="flex gap-2">
                            <input type="number" step="0.001" min="0.001" placeholder="ปริมาณ" value={bomQuantity} onChange={(e) => setBomQuantity(e.target.value)}
                              className="flex-1 border rounded px-2 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" required />
                            <button type="submit" disabled={addingBom} className="bg-blue-600 text-white px-3 py-1.5 rounded text-sm hover:bg-blue-700 disabled:opacity-50 transition">เพิ่ม</button>
                            <button type="button" onClick={() => { setShowAddBom(false); setBomIngredientId(""); setBomSubMenuId(""); setBomQuantity(""); setBomType("ingredient"); }}
                              className="px-3 py-1.5 rounded text-sm text-gray-600 hover:bg-gray-100 transition">ยกเลิก</button>
                          </div>
                        </form>
                      ) : (
                        <button onClick={() => setShowAddBom(true)}
                          className="w-full mt-2 border-2 border-dashed border-gray-200 rounded-lg py-2 text-sm text-gray-500 hover:border-blue-300 hover:text-blue-600 transition">
                          + เพิ่มส่วนประกอบ BOM
                        </button>
                      )}
                    </>
                  )
                )}

                {/* ── Add-ons Tab ── */}
                {activeTab === "addons" && (
                  addonsLoading ? <p className="text-sm text-gray-400 text-center py-6">กำลังโหลด...</p> : (
                    <>
                      {addons.length === 0 ? <p className="text-sm text-gray-400 mb-3">ยังไม่มี Topping</p> : (
                        <div className="space-y-2 mb-4">
                          {addons.map((addon) => (
                            <div key={addon.id} className="flex items-center justify-between bg-gray-50 rounded px-3 py-2">
                              <div>
                                <p className="text-sm font-medium text-gray-700">
                                  {addon.ingredient?.name ?? ingredientName(addon.ingredient_id)}
                                  {addon.extra_price > 0 && <span className="text-orange-600 ml-1">+฿{addon.extra_price}</span>}
                                </p>
                                <p className="text-xs text-gray-400">
                                  {addon.quantity} {addon.ingredient?.unit ?? ingredientUnit(addon.ingredient_id)}
                                  {!addon.is_available && <span className="ml-2 text-red-400">ไม่พร้อมขาย</span>}
                                </p>
                              </div>
                              <button onClick={() => handleDeleteAddon(addon.id)} className="text-xs text-red-500 hover:text-red-700">ลบ</button>
                            </div>
                          ))}
                        </div>
                      )}
                      {showAddAddon ? (
                        <form onSubmit={handleAddAddon} className="border-t pt-4 space-y-3">
                          <p className="text-sm font-medium text-gray-600">เพิ่ม Topping</p>
                          <select value={addonIngredientId} onChange={(e) => setAddonIngredientId(e.target.value)}
                            className="w-full border rounded px-2 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" required>
                            <option value="">เลือกวัตถุดิบ...</option>
                            {ingredients.map((ing) => <option key={ing.id} value={ing.id}>{ing.name} ({ing.unit})</option>)}
                          </select>
                          <div className="flex gap-2">
                            <input type="number" step="0.001" min="0.001" placeholder="ปริมาณ" value={addonQuantity} onChange={(e) => setAddonQuantity(e.target.value)}
                              className="flex-1 border rounded px-2 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" required />
                            <input type="number" step="0.01" min="0" placeholder="ราคาเพิ่ม (฿)" value={addonExtraPrice} onChange={(e) => setAddonExtraPrice(e.target.value)}
                              className="flex-1 border rounded px-2 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
                          </div>
                          <div className="flex gap-2">
                            <button type="submit" disabled={addingAddon} className="bg-blue-600 text-white px-3 py-1.5 rounded text-sm hover:bg-blue-700 disabled:opacity-50 transition">เพิ่ม</button>
                            <button type="button" onClick={() => { setShowAddAddon(false); setAddonIngredientId(""); setAddonQuantity(""); setAddonExtraPrice(""); }}
                              className="px-3 py-1.5 rounded text-sm text-gray-600 hover:bg-gray-100 transition">ยกเลิก</button>
                          </div>
                        </form>
                      ) : (
                        <button onClick={() => setShowAddAddon(true)}
                          className="w-full mt-2 border-2 border-dashed border-gray-200 rounded-lg py-2 text-sm text-gray-500 hover:border-blue-300 hover:text-blue-600 transition">
                          + เพิ่ม Topping
                        </button>
                      )}
                    </>
                  )
                )}

                {/* ── Choices Tab ── */}
                {activeTab === "choices" && (
                  choicesLoading ? <p className="text-sm text-gray-400 text-center py-6">กำลังโหลด...</p> : (
                    <>
                      {choices.length === 0 ? <p className="text-sm text-gray-400 mb-3">ยังไม่มีตัวเลือก</p> : (
                        <div className="space-y-4 mb-4">
                          {choices.map((group) => (
                            <div key={group.id} className="border rounded-lg overflow-hidden">
                              <div className="flex items-center justify-between bg-gray-50 px-3 py-2">
                                <div className="flex items-center gap-2 flex-wrap">
                                  <p className="text-sm font-semibold text-gray-700">{group.name}</p>
                                  <span className={`text-xs px-1.5 py-0.5 rounded ${group.is_required ? "bg-red-100 text-red-600" : "bg-gray-100 text-gray-500"}`}>
                                    {group.is_required ? "บังคับ" : "ไม่บังคับ"}
                                  </span>
                                  <span className="text-xs text-gray-400">เลือก {group.min_choices}–{group.max_choices}</span>
                                </div>
                                <div className="flex gap-1">
                                  <button onClick={() => { setAddingOptionToGroup(group.id); setOptionIngredientId(""); setOptionQuantity(""); setOptionExtraPrice(""); }}
                                    className="text-xs text-blue-600 hover:text-blue-800 px-2">+ ตัวเลือก</button>
                                  <button onClick={() => handleDeleteGroup(group.id)} className="text-xs text-red-500 hover:text-red-700 px-2">ลบกลุ่ม</button>
                                </div>
                              </div>
                              {(group.options ?? []).length > 0 && (
                                <div className="divide-y">
                                  {(group.options ?? []).map((opt) => (
                                    <div key={opt.id} className="flex items-center justify-between px-3 py-1.5">
                                      <div>
                                        <span className="text-sm text-gray-700">{opt.ingredient?.name ?? ingredientName(opt.ingredient_id)}</span>
                                        {opt.extra_price > 0 && <span className="text-xs text-orange-600 ml-2">+฿{opt.extra_price}</span>}
                                        <span className="text-xs text-gray-400 ml-2">{opt.quantity} {opt.ingredient?.unit ?? ingredientUnit(opt.ingredient_id)}</span>
                                      </div>
                                      <button onClick={() => handleDeleteOption(group.id, opt.id)} className="text-xs text-red-400 hover:text-red-600">ลบ</button>
                                    </div>
                                  ))}
                                </div>
                              )}
                              {addingOptionToGroup === group.id && (
                                <form onSubmit={handleAddOption} className="border-t bg-blue-50 px-3 py-3 space-y-2">
                                  <p className="text-xs font-medium text-blue-700">เพิ่มตัวเลือกใน &ldquo;{group.name}&rdquo;</p>
                                  <select value={optionIngredientId} onChange={(e) => setOptionIngredientId(e.target.value)}
                                    className="w-full border rounded px-2 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" required>
                                    <option value="">เลือกวัตถุดิบ...</option>
                                    {ingredients.map((ing) => <option key={ing.id} value={ing.id}>{ing.name} ({ing.unit})</option>)}
                                  </select>
                                  <div className="flex gap-2">
                                    <input type="number" step="0.001" min="0.001" placeholder="ปริมาณ" value={optionQuantity} onChange={(e) => setOptionQuantity(e.target.value)}
                                      className="flex-1 border rounded px-2 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" required />
                                    <input type="number" step="0.01" min="0" placeholder="ราคาเพิ่ม (฿)" value={optionExtraPrice} onChange={(e) => setOptionExtraPrice(e.target.value)}
                                      className="flex-1 border rounded px-2 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
                                  </div>
                                  <div className="flex gap-2">
                                    <button type="submit" disabled={addingOption} className="bg-blue-600 text-white px-3 py-1.5 rounded text-sm hover:bg-blue-700 disabled:opacity-50 transition">เพิ่ม</button>
                                    <button type="button" onClick={() => setAddingOptionToGroup(null)} className="px-3 py-1.5 rounded text-sm text-gray-600 hover:bg-gray-100 transition">ยกเลิก</button>
                                  </div>
                                </form>
                              )}
                            </div>
                          ))}
                        </div>
                      )}
                      {showAddGroup ? (
                        <form onSubmit={handleAddGroup} className="border-t pt-4 space-y-3">
                          <p className="text-sm font-medium text-gray-600">เพิ่มกลุ่มตัวเลือก</p>
                          <input type="text" placeholder="ชื่อกลุ่ม เช่น ระดับความเผ็ด" value={groupName} onChange={(e) => setGroupName(e.target.value)}
                            className="w-full border rounded px-2 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" required />
                          <div className="flex gap-3 flex-wrap">
                            <label className="flex items-center gap-1.5 text-sm text-gray-600 cursor-pointer">
                              <input type="checkbox" checked={groupRequired} onChange={(e) => setGroupRequired(e.target.checked)} className="rounded" />
                              บังคับเลือก
                            </label>
                            <div className="flex items-center gap-2 text-sm">
                              <span className="text-gray-500">เลือกขั้นต่ำ</span>
                              <input type="number" min="0" max="99" value={groupMin} onChange={(e) => setGroupMin(e.target.value)}
                                className="w-14 border rounded px-2 py-1 text-sm text-center focus:outline-none focus:ring-2 focus:ring-blue-500" />
                              <span className="text-gray-500">สูงสุด</span>
                              <input type="number" min="1" max="99" value={groupMax} onChange={(e) => setGroupMax(e.target.value)}
                                className="w-14 border rounded px-2 py-1 text-sm text-center focus:outline-none focus:ring-2 focus:ring-blue-500" />
                            </div>
                          </div>
                          <div className="flex gap-2">
                            <button type="submit" disabled={addingGroup} className="bg-blue-600 text-white px-3 py-1.5 rounded text-sm hover:bg-blue-700 disabled:opacity-50 transition">เพิ่มกลุ่ม</button>
                            <button type="button" onClick={() => { setShowAddGroup(false); setGroupName(""); }} className="px-3 py-1.5 rounded text-sm text-gray-600 hover:bg-gray-100 transition">ยกเลิก</button>
                          </div>
                        </form>
                      ) : (
                        <button onClick={() => setShowAddGroup(true)}
                          className="w-full mt-2 border-2 border-dashed border-gray-200 rounded-lg py-2 text-sm text-gray-500 hover:border-blue-300 hover:text-blue-600 transition">
                          + เพิ่มกลุ่มตัวเลือก
                        </button>
                      )}
                    </>
                  )
                )}

                {/* ── Portions Tab ── */}
                {activeTab === "portions" && (
                  portionsLoading ? <p className="text-sm text-gray-400 text-center py-6">กำลังโหลด...</p> : (
                    <>
                      {portions.length === 0 ? <p className="text-sm text-gray-400 mb-3">ยังไม่มีขนาด</p> : (
                        <div className="space-y-2 mb-4">
                          {portions.map((portion) => (
                            <div key={portion.id} className="flex items-center justify-between bg-gray-50 rounded px-3 py-2">
                              <div>
                                <div className="flex items-center gap-2">
                                  <p className="text-sm font-medium text-gray-700">{portion.name}</p>
                                  {portion.is_default && <span className="text-xs bg-green-100 text-green-600 px-1.5 py-0.5 rounded">ค่าเริ่มต้น</span>}
                                </div>
                                <p className="text-xs text-gray-400">
                                  ×{portion.quantity_multiplier}
                                  {portion.extra_price > 0 && <span className="ml-2 text-orange-600">+฿{portion.extra_price}</span>}
                                </p>
                              </div>
                              <button onClick={() => handleDeletePortion(portion.id)} className="text-xs text-red-500 hover:text-red-700">ลบ</button>
                            </div>
                          ))}
                        </div>
                      )}
                      {showAddPortion ? (
                        <form onSubmit={handleAddPortion} className="border-t pt-4 space-y-3">
                          <p className="text-sm font-medium text-gray-600">เพิ่มขนาด</p>
                          <input type="text" placeholder="ชื่อขนาด เช่น พิเศษ, ธรรมดา" value={portionName} onChange={(e) => setPortionName(e.target.value)}
                            className="w-full border rounded px-2 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" required />
                          <div className="flex gap-2">
                            <div className="flex-1">
                              <label className="block text-xs text-gray-500 mb-1">ตัวคูณปริมาณ</label>
                              <input type="number" step="0.01" min="0.1" placeholder="1.0" value={portionMultiplier} onChange={(e) => setPortionMultiplier(e.target.value)}
                                className="w-full border rounded px-2 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
                            </div>
                            <div className="flex-1">
                              <label className="block text-xs text-gray-500 mb-1">ราคาเพิ่ม (฿)</label>
                              <input type="number" step="0.01" min="0" placeholder="0" value={portionExtraPrice} onChange={(e) => setPortionExtraPrice(e.target.value)}
                                className="w-full border rounded px-2 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
                            </div>
                          </div>
                          <label className="flex items-center gap-1.5 text-sm text-gray-600 cursor-pointer">
                            <input type="checkbox" checked={portionIsDefault} onChange={(e) => setPortionIsDefault(e.target.checked)} className="rounded" />
                            ตั้งเป็นขนาดเริ่มต้น
                          </label>
                          <div className="flex gap-2">
                            <button type="submit" disabled={addingPortion} className="bg-blue-600 text-white px-3 py-1.5 rounded text-sm hover:bg-blue-700 disabled:opacity-50 transition">เพิ่ม</button>
                            <button type="button" onClick={() => { setShowAddPortion(false); setPortionName(""); setPortionMultiplier("1.0"); setPortionExtraPrice("0"); setPortionIsDefault(false); }}
                              className="px-3 py-1.5 rounded text-sm text-gray-600 hover:bg-gray-100 transition">ยกเลิก</button>
                          </div>
                        </form>
                      ) : (
                        <button onClick={() => setShowAddPortion(true)}
                          className="w-full mt-2 border-2 border-dashed border-gray-200 rounded-lg py-2 text-sm text-gray-500 hover:border-blue-300 hover:text-blue-600 transition">
                          + เพิ่มขนาด
                        </button>
                      )}
                    </>
                  )
                )}
              </div>
            </div>
          </div>
        )}
      </div>

      {/* ─── Create/Edit Modal ─────────────────────────────────── */}
      {showModal && (
        <div className="fixed inset-0 bg-black/40 flex items-center justify-center z-50">
          <div className="bg-white rounded-xl shadow-lg w-full max-w-md mx-4 p-6">
            <h2 className="text-lg font-bold mb-4">{editingMenu ? "แก้ไขเมนู" : "เพิ่มเมนูใหม่"}</h2>
            <form onSubmit={handleSave} className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">ชื่อเมนู *</label>
                <input type="text" value={formName} onChange={(e) => setFormName(e.target.value)}
                  className="w-full border rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" placeholder="เช่น ข้าวผัดกระเพราหมูสับ" required />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">รายละเอียด</label>
                <textarea value={formDescription} onChange={(e) => setFormDescription(e.target.value)}
                  className="w-full border rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" rows={2} placeholder="คำอธิบายเมนู" />
              </div>
              <div className="flex gap-3">
                <div className="flex-1">
                  <label className="block text-sm font-medium text-gray-700 mb-1">ราคา (฿) *</label>
                  <input type="number" step="0.01" min="0" value={formPrice} onChange={(e) => setFormPrice(e.target.value)}
                    className="w-full border rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" placeholder="0.00" required />
                </div>
                <div className="flex-1">
                  <label className="block text-sm font-medium text-gray-700 mb-1">หมวดหมู่ *</label>
                  <input type="text" value={formCategory} onChange={(e) => setFormCategory(e.target.value)}
                    className="w-full border rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" placeholder="เช่น อาหารจานเดียว" required />
                </div>
              </div>
              <label className="flex items-center gap-2 cursor-pointer">
                <input type="checkbox" id="is_available" checked={formAvailable} onChange={(e) => setFormAvailable(e.target.checked)} className="rounded" />
                <span className="text-sm text-gray-700">เปิดขาย</span>
              </label>
              <div className="flex justify-end gap-2 pt-2">
                <button type="button" onClick={() => setShowModal(false)} className="px-4 py-2 text-sm text-gray-600 hover:text-gray-800 transition">ยกเลิก</button>
                <button type="submit" disabled={saving} className="bg-blue-600 text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-blue-700 disabled:opacity-50 transition">
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
              <button onClick={() => setDeletingId(null)} className="px-4 py-2 text-sm text-gray-600 hover:text-gray-800 transition">ยกเลิก</button>
              <button onClick={handleDelete} className="bg-red-600 text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-red-700 transition">ลบเมนู</button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
