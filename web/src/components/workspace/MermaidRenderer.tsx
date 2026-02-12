"use client";

import { useEffect, useRef, useState, useId } from "react";
import mermaid from "mermaid";

interface MermaidRendererProps {
  code: string;
  colorScheme?: string;
}

export function MermaidRenderer({ code, colorScheme }: MermaidRendererProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const [error, setError] = useState<string | null>(null);
  const uniqueId = useId().replace(/:/g, "_");

  useEffect(() => {
    if (!code.trim() || !containerRef.current) return;

    const theme = colorScheme === "drawio" ? "default" : "neutral";

    mermaid.initialize({
      startOnLoad: false,
      theme,
      securityLevel: "loose",
      fontFamily: "inherit",
    });

    let cancelled = false;

    (async () => {
      try {
        const { svg } = await mermaid.render(`mermaid_${uniqueId}`, code.trim());
        if (!cancelled && containerRef.current) {
          containerRef.current.innerHTML = svg;
          setError(null);
        }
      } catch (err) {
        if (!cancelled) {
          setError(
            err instanceof Error ? err.message : "Mermaid 图表渲染失败"
          );
          if (containerRef.current) {
            containerRef.current.innerHTML = "";
          }
        }
      }
    })();

    return () => {
      cancelled = true;
    };
  }, [code, colorScheme, uniqueId]);

  if (error) {
    return (
      <div className="flex h-full min-h-[400px] items-center justify-center rounded-md border border-destructive/50 bg-destructive/5 p-4">
        <div className="max-w-full space-y-2 text-center">
          <p className="text-sm font-medium text-destructive">
            Mermaid 渲染错误
          </p>
          <p className="whitespace-pre-wrap break-words text-xs text-muted-foreground">
            {error}
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="flex min-h-[400px] items-center justify-center rounded-md border bg-white p-4">
      <div ref={containerRef} className="max-w-full overflow-auto" />
    </div>
  );
}
