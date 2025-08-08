"use client";
import Image from "next/image";
import { TypographyTitle, TypographyH1, TypographyH2, TypographyH3, TypographyH4, TypographyP, TypographyInlineCode } from "@/components/ui/typography";
import Logo from "@/components/logo";
import { useTheme } from "next-themes";

export default function Home() {
  const { theme } = useTheme()
  return (
    <main>
      <div className="flex flex-col items-center min-h-screen gap-2">
        <Logo
          theme={theme}
          className="size-24 xl:size-32"
        />
        <TypographyTitle>Gomegle</TypographyTitle>
        <TypographyP>
          Chat with strangers, anonymously.
        </TypographyP>
        <TypographyInlineCode>
          ssh gomegle.sh
        </TypographyInlineCode>
      </div>
    </main>
  );
}
