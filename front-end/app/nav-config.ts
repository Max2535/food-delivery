export interface NavItem {
  label: string;
  href: string;
  permissions: string[];
}

export interface NavGroup {
  label: string;
  permissions: string[];
  items: NavItem[];
}

/**
 * ดึง menu config จาก backend (กรองตาม permissions ของ user แล้ว)
 */
export async function fetchMenuConfig(
  accessToken: string
): Promise<NavGroup[]> {
  const res = await fetch("/api/auth/menu-config", {
    headers: { Authorization: `Bearer ${accessToken}` },
  });
  if (!res.ok) return [];
  const data = await res.json();
  return data.menu ?? [];
}
