import { NextRequest, NextResponse } from "next/server";

const API_URL = process.env.API_URL || "http://localhost:8080";

export async function GET(req: NextRequest) {
  const authHeader = req.headers.get("authorization") || "";

  const res = await fetch(`${API_URL}/v1/auth/groups`, {
    headers: {
      "Content-Type": "application/json",
      Authorization: authHeader,
    },
  });

  const data = await res.json().catch(() => null);
  return NextResponse.json(data ?? { message: "Unknown error" }, { status: res.status });
}

export async function POST(req: NextRequest) {
  const authHeader = req.headers.get("authorization") || "";
  const body = await req.json();

  const res = await fetch(`${API_URL}/v1/auth/group`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: authHeader,
    },
    body: JSON.stringify(body),
  });

  const data = await res.json().catch(() => null);
  return NextResponse.json(data ?? { message: "Unknown error" }, { status: res.status });
}

export async function PUT(req: NextRequest) {
  const authHeader = req.headers.get("authorization") || "";
  const body = await req.json();

  const res = await fetch(`${API_URL}/v1/auth/group`, {
    method: "PUT",
    headers: {
      "Content-Type": "application/json",
      Authorization: authHeader,
    },
    body: JSON.stringify(body),
  });

  const data = await res.json().catch(() => null);
  return NextResponse.json(data ?? { message: "Unknown error" }, { status: res.status });
}

export async function DELETE(req: NextRequest) {
  const authHeader = req.headers.get("authorization") || "";
  const { id } = await req.json();

  const res = await fetch(`${API_URL}/v1/auth/group/${id}`, {
    method: "DELETE",
    headers: {
      "Content-Type": "application/json",
      Authorization: authHeader,
    },
  });

  const data = await res.json().catch(() => null);
  return NextResponse.json(data ?? { message: "Unknown error" }, { status: res.status });
}