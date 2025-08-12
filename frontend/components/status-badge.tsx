"use client";

import { useEffect, useState } from "react";
import { Badge } from "@/components/ui/badge";

export default function StatusBadge() {
  const [n, setN] = useState<number | null>(null);
  const [ok, setOk] = useState(true);

  useEffect(() => {
    let alive = true;

    async function tick() {
      try {
        const res = await fetch("/api/status", { cache: "no-store" });
        setOk(res.ok);
        const data = await res.json();
        if (alive) setN(data.active ?? 0);
      } catch {
        if (alive) setN(null);
      }
    }

    tick();
    const id = setInterval(tick, 5000); // 5s
    return () => { alive = false; clearInterval(id); };
  }, []);

  return (
    <Badge
      variant="secondary"
      className="bg-zinc-800 text-zinc-200 flex items-center gap-2"
      title="Users currently chatting"
    >
      {
        ok ? (
          <span className="relative block h-2 w-2 rounded-full bg-emerald-400">
            <span className="absolute inset-0 rounded-full animate-ping bg-emerald-400/70" />
          </span>
        ) : (
          <span className="relative block h-2 w-2 rounded-full bg-red-400">
            <span className="absolute inset-0 rounded-full animate-ping bg-red-400/70" />
          </span>
        )
      }
      {
        ok ? "Online: " : "Offline"
      }
      {
        ok && n !== null ? (
          <>{n.toLocaleString()}</>
        ) : (
          <></>
        )
      }
    </Badge>
  );
}
