import { NextRequest, NextResponse } from "next/server";

const API_URL = process.env.API_URL || "http://localhost:8080";

export async function POST(req: NextRequest) {
  const body = await req.json();

  const res = await fetch(`${API_URL}/v1/auth/forgot-password`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });

  const data = await res.json().catch(() => null);
  return NextResponse.json(data ?? { message: "Unknown error" }, { status: res.status });
}
