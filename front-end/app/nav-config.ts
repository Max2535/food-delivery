export interface NavItem {
  label: string;
  href: string;
  roles: string[]; // ถ้าว่าง = ทุก role เข้าได้
}

export interface NavGroup {
  label: string;
  roles: string[]; // ถ้าว่าง = ทุก role เข้าได้
  items: NavItem[];
}

export const NAV_GROUPS: NavGroup[] = [
  {
    label: "Auth",
    roles: ["admin", "manager"],
    items: [
      { label: "กลุ่ม", href: "/dashboard/groups", roles: ["admin", "manager"] },
      { label: "สิทธิ์", href: "/dashboard/roles", roles: ["admin", "manager"] },
      { label: "ผู้ใช้", href: "/dashboard/users", roles: ["admin", "manager"] },
    ],
  },
  {
    label: "Catalog",
    roles: [],
    items: [
      { label: "เมนู", href: "/dashboard/menus", roles: ["admin", "manager"] },
      { label: "วัตถุดิบ (BOM)", href: "/dashboard/ingredients", roles: ["admin", "manager"] },
    ],
  },
  {
    label: "Kitchen",
    roles: [],
    items: [
      { label: "ครัว", href: "/dashboard/kitchen", roles: ["admin", "manager"] },
    ],
  },
  {
    label: "Order",
    roles: [],
    items: [
      { label: "ออเดอร์", href: "/dashboard/orders", roles: ["admin", "manager"] },
      { label: "คิว", href: "/dashboard/queue", roles: ["admin", "manager"] },
      { label: "สั่งอาหาร", href: "/dashboard/orders/create", roles: ["user"] },
    ],
  },
  {
    label: "Inventory",
    roles: [],
    items: [
      { label: "รายการเดินสต๊อก", href: "/dashboard/inventory", roles: ["admin", "manager"] },
    ],
  },
];

/**
 * กรอง nav groups ตาม roles ของ user
 * - ถ้า group.roles ว่าง → แสดงเสมอ
 * - ถ้า user มี role "admin" → แสดงทุก group
 * - กรอง items ภายใน group ด้วย
 * - ถ้า group ไม่มี item เหลือ → ไม่แสดง group
 */
export function filterNavGroupsByRoles(
  groups: NavGroup[],
  userRoles: string[]
): NavGroup[] {
  const isAdmin = userRoles.includes("admin");

  return groups
    .filter(
      (group) =>
        isAdmin ||
        group.roles.length === 0 ||
        group.roles.some((r) => userRoles.includes(r))
    )
    .map((group) => ({
      ...group,
      items: isAdmin
        ? group.items
        : group.items.filter(
            (item) =>
              item.roles.length === 0 ||
              item.roles.some((r) => userRoles.includes(r))
          ),
    }))
    .filter((group) => group.items.length > 0);
}
