import { NextRequest, NextResponse } from "next/server";

const API_URL = process.env.API_URL || "http://localhost:8080";

export async function DELETE(
  req: NextRequest,
  { params }: { params: Promise<{ id: string; portion_id: string }> }
) {
  const { id, portion_id } = await params;
  const authHeader = req.headers.get("authorization") || "";

  const res = await fetch(
    `${API_URL}/v1/catalog/menus/${id}/portions/${portion_id}`,
    {
      method: "DELETE",
      headers: { Authorization: authHeader },
    }
  );

  const data = await res.json().catch(() => null);
  return NextResponse.json(data ?? { message: "Unknown error" }, { status: res.status });
}
