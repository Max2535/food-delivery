export interface NavItem {
  label: string;
  href: string;
  roles: string[]; // ถ้าว่าง = ทุก role เข้าได้
}

export const NAV_ITEMS: NavItem[] = [
  {
    label: "Dashboard",
    href: "/dashboard",
    roles: [], // ทุกคนเข้าได้
  },
  {
    label: "กลุ่ม",
    href: "/dashboard/groups",
    roles: ["admin", "manager"],
  },
];

/**
 * กรองเมนูตาม roles ของ user
 * - ถ้า item.roles ว่าง → แสดงเสมอ
 * - ถ้า user มี role "admin" → แสดงทุกเมนู
 * - ถ้า user มี role ตรงกับ item.roles อย่างน้อย 1 ตัว → แสดง
 */
export function filterNavByRoles(items: NavItem[], userRoles: string[]): NavItem[] {
  if (userRoles.includes("admin")) return items;

  return items.filter(
    (item) =>
      item.roles.length === 0 ||
      item.roles.some((r) => userRoles.includes(r))
  );
}
