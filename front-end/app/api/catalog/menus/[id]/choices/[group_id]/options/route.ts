import { NextRequest, NextResponse } from "next/server";

const API_URL = process.env.API_URL || "http://localhost:8080";

export async function POST(
  req: NextRequest,
  { params }: { params: Promise<{ id: string; group_id: string }> }
) {
  const { id, group_id } = await params;
  const authHeader = req.headers.get("authorization") || "";
  const body = await req.json();

  const res = await fetch(
    `${API_URL}/v1/catalog/menus/${id}/choices/${group_id}/options`,
    {
      method: "POST",
      headers: { "Content-Type": "application/json", Authorization: authHeader },
      body: JSON.stringify(body),
    }
  );

  const data = await res.json().catch(() => null);
  return NextResponse.json(data ?? { message: "Unknown error" }, { status: res.status });
}
