"use client";

import React, { useEffect, useMemo, useState } from "react";
import { motion } from "framer-motion";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";
import { Separator } from "@/components/ui/separator";
import { Github, GitBranchPlus, Terminal, Copy, Check, ExternalLink, Shield, Zap, Globe, Server, Lock, Linkedin } from "lucide-react";
import StatusBadge from "@/components/status-badge";
import Logo from "@/components/logo";

/**
 * Drop this file in `app/page.tsx` (App Router) or `pages/index.tsx` (Pages Router).
 * Tailwind and shadcn/ui should already be configured in your project.
 * Replace the LINKS below with your real URLs.
 */

const LINKS = {
  ssh: "ssh gomegle.sh",
  github: "https://github.com/johan253/gomegle",
  contribute: "https://github.com/johan253/gomegle/issues",
  portfolio: "https://johanhernandez.com",
  linkedin: "https://linkedin.com/in/johan253",
  tools: {
    next: "https://nextjs.org/",
    tailwind: "https://tailwindcss.com/",
    shadcn: "https://ui.shadcn.com/",
    go: "https://go.dev/",
    redis: "https://redis.io/",
  },
};

export default function GomegleLanding() {
  const [copied, setCopied] = useState(false);
  const [cursor, setCursor] = useState(true);

  // Minimal typing animation for the terminal line
  const full = useMemo(() => LINKS.ssh, []);
  const [typed, setTyped] = useState("");
  useEffect(() => {
    let i = 0;
    const t = setInterval(() => {
      setTyped(full.slice(0, i + 1));
      i++;
      if (i >= full.length) {
        clearInterval(t);
      }
    }, 55);
    return () => clearInterval(t);
  }, [full]);

  useEffect(() => {
    const blinker = setInterval(() => setCursor((c) => !c), 500);
    return () => clearInterval(blinker);
  }, []);

  const copySSH = async () => {
    try {
      await navigator.clipboard.writeText(LINKS.ssh);
      setCopied(true);
      setTimeout(() => setCopied(false), 1200);
    } catch { }
  };

  return (
    <TooltipProvider>
      <main className="relative min-h-screen overflow-hidden bg-black text-zinc-100">
        {/* Subtle grid background */}
        <div
          aria-hidden
          className="pointer-events-none absolute inset-0 bg-[radial-gradient(circle_at_1px_1px,rgba(255,255,255,0.06)_1px,transparent_1px)] [background-size:24px_24px]"
        />

        {/* Glow accents */}
        <div aria-hidden className="absolute -top-32 left-1/2 h-64 w-[40rem] -translate-x-1/2 rounded-full bg-fuchsia-500/10 blur-3xl" />
        <div aria-hidden className="absolute bottom-[-10rem] right-[-10rem] h-96 w-96 rounded-full bg-indigo-500/10 blur-3xl" />

        {/* Nav */}
        <header className="mx-auto flex w-full max-w-6xl items-center justify-between px-6 py-6">
          <div className="flex items-center gap-3">
            <div className="grid h-9 w-9 place-items-center rounded-xl border border-zinc-800 bg-zinc-900/60 shadow-sm">
              <Logo className="h-5 w-5" />
            </div>
            <span className="font-mono text-lg tracking-tight">Gomegle</span>
          </div>
          <nav className="hidden items-center gap-2 sm:flex">
            <StatusBadge />
            <Button variant="ghost" asChild>
              <a href={LINKS.github} target="_blank" rel="noreferrer">
                <Github className="mr-2 h-4 w-4" /> GitHub
              </a>
            </Button>
            <Button variant="ghost" asChild>
              <a href={LINKS.portfolio} target="_blank" rel="noreferrer">
                <Globe className="mr-2 h-4 w-4" /> Portfolio
              </a>
            </Button>
          </nav>
        </header>

        {/* Hero */}
        <section className="mx-auto grid w-full max-w-6xl grid-cols-1 items-center gap-10 px-6 pb-24 pt-6 md:grid-cols-2">
          <div>
            <motion.h1
              initial={{ opacity: 0, y: 10 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.6 }}
              className="bg-gradient-to-b from-zinc-50 to-zinc-400 bg-clip-text text-5xl font-extrabold leading-tight text-transparent md:text-6xl"
            >
              Anonymous terminal chat over SSH.
            </motion.h1>
            <motion.p
              initial={{ opacity: 0, y: 10 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.6, delay: 0.1 }}
              className="mt-4 max-w-prose text-zinc-300"
            >
              Gomegle pairs you with a stranger in a secure PTY. No browser. No frills. Just raw, low-latency conversation—right from your shell.
            </motion.p>

            <div className="mt-6 flex flex-wrap items-center gap-3">
              <Tooltip>
                <TooltipTrigger asChild>
                  <Button size="lg" onClick={copySSH} className="font-mono">
                    {copied ? (
                      <>
                        <Check className="mr-2 h-4 w-4" /> Copied
                      </>
                    ) : (
                      <>
                        <Copy className="mr-2 h-4 w-4" /> {LINKS.ssh}
                      </>
                    )}
                  </Button>
                </TooltipTrigger>
                <TooltipContent>
                  <p>Copy connect command</p>
                </TooltipContent>
              </Tooltip>

              <Button variant="outline" size="lg" asChild>
                <a href={"ssh://gomegle.sh"} target="_blank" rel="noreferrer">
                  <ExternalLink className="mr-2 h-4 w-4" /> Connect Now
                </a>
              </Button>
            </div>

            <div className="mt-6 flex flex-wrap items-center gap-2 text-sm text-zinc-400">
              <Badge variant="secondary" className="bg-zinc-800 text-zinc-200">Open source</Badge>
              <span>•</span>
              <span>Minimal logs</span>
              <span>•</span>
              <span>Fast matchmaking</span>
            </div>
          </div>

          {/* Terminal mock */}
          <motion.div
            initial={{ opacity: 0, scale: 0.98 }}
            animate={{ opacity: 1, scale: 1 }}
            transition={{ duration: 0.5, delay: 0.1 }}
          >
            <Card className="border-zinc-800 bg-zinc-900/50 shadow-xl backdrop-blur">
              <CardHeader className="border-b border-zinc-800">
                <CardTitle className="flex items-center gap-2 text-sm text-zinc-300">
                  <span className="inline-flex h-2.5 w-2.5 rounded-full bg-red-500" />
                  <span className="inline-flex h-2.5 w-2.5 rounded-full bg-yellow-500" />
                  <span className="inline-flex h-2.5 w-2.5 rounded-full bg-green-500" />
                  <span className="ml-2 font-mono">/bin/zsh — ssh gomegle.sh</span>
                </CardTitle>
              </CardHeader>
              <CardContent>
                <pre className="scrollbar-none h-64 overflow-hidden rounded-lg bg-black/60 p-4 font-mono text-sm leading-7 text-zinc-200">
                  {`$ ${typed}${cursor ? "█" : " "}
Connecting to gomegle.sh...
Match found. Say hi!\n`}
                  <span className="text-zinc-500">Stranger:</span> hey<br />
                  <span className="text-sky-400">You:</span> what&apos;s your favorite shell?<br />
                  <span className="text-zinc-500">Stranger:</span> zsh with starship 😎<br />
                  <span className="text-sky-400">You:</span> nice<br />
                </pre>
                <div className="mt-3 flex items-center gap-2 text-xs text-zinc-500">
                  <Lock className="h-3.5 w-3.5" /> Ephemeral sessions • No accounts required
                </div>
              </CardContent>
            </Card>
          </motion.div>
        </section>

        {/* How it works */}
        <section className="mx-auto w-full max-w-6xl px-6 pb-12">
          <h2 className="mb-6 text-xl font-semibold text-zinc-200">How it works</h2>
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
            {[
              { icon: Terminal, title: "Connect", text: `Run \`${LINKS.ssh}\` from any machine with SSH.` },
              { icon: Server, title: "Queue", text: "We'll pop you into a matchmaking queue." },
              { icon: Zap, title: "Chat", text: "You're paired into a TTY room. Low-latency, pure text." },
              { icon: Shield, title: "Leave", text: "Close your session. Nothing to clean up." },
            ].map((f, i) => (
              <motion.div key={f.title} initial={{ opacity: 0, y: 8 }} whileInView={{ opacity: 1, y: 0 }} transition={{ duration: 0.4, delay: i * 0.05 }} viewport={{ once: true }}>
                <Card className="border-zinc-800 bg-zinc-900/50">
                  <CardHeader className="flex flex-row items-center gap-3">
                    <f.icon className="h-5 w-5" />
                    <CardTitle className="text-base">{f.title}</CardTitle>
                  </CardHeader>
                  <CardContent className="text-sm text-zinc-400">{f.text}</CardContent>
                </Card>
              </motion.div>
            ))}
          </div>
        </section>

        <Separator className="mx-auto my-8 w-[calc(100%-3rem)] bg-zinc-800" />

        {/* Stack & Links */}
        <section className="mx-auto w-full max-w-6xl px-6 pb-24">
          <h2 className="mb-4 text-xl font-semibold text-zinc-200">Built with</h2>
          <div className="flex flex-wrap items-center gap-2">
            <TechBadge name="Next.js" href={LINKS.tools.next} />
            <TechBadge name="TailwindCSS" href={LINKS.tools.tailwind} />
            <TechBadge name="shadcn/ui" href={LINKS.tools.shadcn} />
            <TechBadge name="Go (backend)" href={LINKS.tools.go} />
            <TechBadge name="Redis (matchmaking)" href={LINKS.tools.redis} />
          </div>

          <div className="mt-8 flex flex-wrap gap-3">
            <Button asChild>
              <a href={LINKS.contribute} target="_blank" rel="noreferrer">
                <GitBranchPlus className="mr-2 h-4 w-4" /> Contribute
              </a>
            </Button>
            <Button variant="outline" asChild>
              <a href={LINKS.linkedin} target="_blank" rel="noreferrer">
                <Linkedin className="mr-2 h-4 w-4" /> Creator
              </a>
            </Button>
          </div>
        </section>

        {/* Footer */}
        <footer className="mx-auto w-full max-w-6xl px-6 pb-10 text-sm text-zinc-500">
          <div className="flex flex-col items-start justify-between gap-2 sm:flex-row sm:items-center">
            <p>© {new Date().getFullYear()} Gomegle. Built for the terminal kids.</p>
            <p className="text-zinc-600">ssh gomegle.sh</p>
          </div>
        </footer>
      </main>
    </TooltipProvider>
  );
}

function TechBadge({ name, href }: { name: string; href: string }) {
  return (
    <a href={href} target="_blank" rel="noreferrer" className="group">
      <Badge
        variant="secondary"
        className="inline-flex cursor-pointer items-center gap-2 border border-zinc-800 bg-zinc-900/60 text-zinc-200 transition hover:translate-y-[-2px] hover:border-zinc-700 hover:bg-zinc-900"
      >
        {name}
      </Badge>
    </a>
  );
}
