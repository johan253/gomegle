import Redis from "ioredis";

declare global {
  var __redis: Redis | undefined;
}

export function getRedis() {
  if (!global.__redis) {
    const url = process.env.REDIS_URL
    if (!url) throw new Error("REDIS_URL is not defined in environment variables");
    try {
      global.__redis = new Redis(url, {
        maxRetriesPerRequest: 2,
        enableAutoPipelining: true,
        lazyConnect: false,
      });
    } catch (err) {
      console.error("Error connecting to Redis:", err);
    }
  }
  return global.__redis;
}
