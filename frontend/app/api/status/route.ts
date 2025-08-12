import { NextResponse } from "next/server";
import { getRedis } from "@/lib/redis";

export const dynamic = "force-dynamic"; // avoid static caching

export async function GET() {
  try {
    const redis = getRedis();
    if (!redis) {
      console.error("Redis client is not initialized");
      return NextResponse.json({ active: 0 }, { status: 500 });
    }
    const val = await redis.get("active"); // string or null
    const count = val ? parseInt(val, 10) : 0;

    // Set a tiny cache so fast refreshes don’t stampede Redis (optional)
    return NextResponse.json(
      { active: Number.isFinite(count) ? count : 0 },
      {
        headers: {
          "Cache-Control": "public, max-age=5, s-maxage=5, stale-while-revalidate=30",
        },
      }
    );
  } catch (err) {
    // Don’t leak internals; just return 0 and a 200 to keep UI simple
    console.error("Error fetching active count:", err);
    return NextResponse.json({ active: 0 }, { status: 500 });
  }
}
