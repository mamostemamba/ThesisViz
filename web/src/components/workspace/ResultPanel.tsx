"use client";

import { useState, useCallback, useRef } from "react";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  Copy,
  Check,
  RefreshCw,
  Download,
  ExternalLink,
  FileCode,
} from "lucide-react";
import { MermaidRenderer } from "./MermaidRenderer";
import { DiffViewer } from "./DiffViewer";
import { useSettingsStore } from "@/stores/useSettingsStore";
import { exportTeX } from "@/lib/api";

interface ResultPanelProps {
  code: string;
  format: string;
  explanation: string;
  imageUrl: string;
  reviewPassed: boolean;
  reviewRounds: number;
  onRefine: (modification: string) => void;
  isRefining: boolean;
  parentCode?: string;
}

export function ResultPanel({
  code,
  format,
  explanation,
  imageUrl,
  reviewPassed,
  reviewRounds,
  onRefine,
  isRefining,
  parentCode,
}: ResultPanelProps) {
  const colorScheme = useSettingsStore((s) => s.colorScheme);
  const language = useSettingsStore((s) => s.language);
  const [modification, setModification] = useState("");
  const [copied, setCopied] = useState<string | null>(null);
  const mermaidContainerRef = useRef<HTMLDivElement>(null);

  const handleCopy = useCallback(
    async (label: string, text: string) => {
      await navigator.clipboard.writeText(text);
      setCopied(label);
      setTimeout(() => setCopied(null), 2000);
    },
    []
  );

  const handleCopyCode = useCallback(() => {
    handleCopy("code", code);
  }, [code, handleCopy]);

  const handleCopyForOverleaf = useCallback(async () => {
    try {
      const result = await exportTeX({ code, language, color_scheme: colorScheme });
      handleCopy("overleaf", result.tex);
    } catch {
      handleCopy("overleaf", code);
    }
  }, [code, language, colorScheme, handleCopy]);

  const handleDownloadPNG = useCallback(async () => {
    if (format === "mermaid") {
      // For Mermaid, convert SVG to PNG via canvas
      const svgEl = mermaidContainerRef.current?.querySelector("svg");
      if (!svgEl) return;

      const { toPng } = await import("html-to-image");
      const dataUrl = await toPng(svgEl as unknown as HTMLElement, { backgroundColor: "#ffffff" });
      const link = document.createElement("a");
      link.download = "figure.png";
      link.href = dataUrl;
      link.click();
    } else if (imageUrl) {
      // For TikZ/Matplotlib, download the server image
      const response = await fetch(imageUrl);
      const blob = await response.blob();
      const url = URL.createObjectURL(blob);
      const link = document.createElement("a");
      link.download = "figure.png";
      link.href = url;
      link.click();
      URL.revokeObjectURL(url);
    }
  }, [format, imageUrl]);

  const handleDownloadSVG = useCallback(() => {
    const svgEl = mermaidContainerRef.current?.querySelector("svg");
    if (!svgEl) return;
    const svgStr = new XMLSerializer().serializeToString(svgEl);
    const blob = new Blob([svgStr], { type: "image/svg+xml" });
    const url = URL.createObjectURL(blob);
    const link = document.createElement("a");
    link.download = "figure.svg";
    link.href = url;
    link.click();
    URL.revokeObjectURL(url);
  }, []);

  const handleOpenMermaidLive = useCallback(() => {
    const state = JSON.stringify({ code, mermaid: { theme: "default" } });
    const encoded = btoa(unescape(encodeURIComponent(state)));
    window.open(
      `https://mermaid.live/edit#base64:${encoded}`,
      "_blank"
    );
  }, [code]);

  const handleRefine = useCallback(() => {
    if (!modification.trim()) return;
    onRefine(modification);
    setModification("");
  }, [modification, onRefine]);

  return (
    <div className="space-y-4">
      {/* Top: Image + Code/Explanation side by side */}
      <div className="grid gap-4 lg:grid-cols-2">
        {/* Left: Image preview */}
        <div className="space-y-2">
          <div className="flex items-center justify-between">
            <h3 className="text-sm font-semibold">Result</h3>
            <div className="flex items-center gap-2 text-xs text-muted-foreground">
              {reviewPassed ? (
                <span className="text-green-600">Review passed</span>
              ) : (
                <span className="text-yellow-600">
                  Review: {reviewRounds} round(s)
                </span>
              )}
            </div>
          </div>
          {format === "mermaid" && code ? (
            <div ref={mermaidContainerRef}>
              <MermaidRenderer code={code} colorScheme={colorScheme} />
            </div>
          ) : imageUrl ? (
            <a
              href={imageUrl}
              target="_blank"
              rel="noopener noreferrer"
              className="block"
            >
              {/* eslint-disable-next-line @next/next/no-img-element */}
              <img
                src={imageUrl}
                alt="Generated figure"
                className="max-h-[400px] w-full rounded border object-contain"
              />
            </a>
          ) : (
            <div className="flex h-[300px] items-center justify-center rounded border border-dashed text-sm text-muted-foreground">
              No image available
            </div>
          )}

          {/* Export buttons */}
          <div className="flex flex-wrap gap-2">
            <Button variant="outline" size="sm" onClick={handleCopyCode}>
              {copied === "code" ? (
                <Check className="mr-1 h-3 w-3" />
              ) : (
                <Copy className="mr-1 h-3 w-3" />
              )}
              {copied === "code" ? "Copied!" : "Copy Code"}
            </Button>

            {format === "tikz" && (
              <Button
                variant="outline"
                size="sm"
                onClick={handleCopyForOverleaf}
              >
                {copied === "overleaf" ? (
                  <Check className="mr-1 h-3 w-3" />
                ) : (
                  <FileCode className="mr-1 h-3 w-3" />
                )}
                {copied === "overleaf" ? "Copied!" : "Copy for Overleaf"}
              </Button>
            )}

            {(imageUrl || format === "mermaid") && (
              <Button variant="outline" size="sm" onClick={handleDownloadPNG}>
                <Download className="mr-1 h-3 w-3" />
                Download PNG
              </Button>
            )}

            {format === "mermaid" && (
              <>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={handleDownloadSVG}
                >
                  <Download className="mr-1 h-3 w-3" />
                  Download SVG
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={handleOpenMermaidLive}
                >
                  <ExternalLink className="mr-1 h-3 w-3" />
                  Mermaid Live
                </Button>
              </>
            )}
          </div>
        </div>

        {/* Right: Code / Explanation tabs */}
        <div className="space-y-2">
          <Tabs defaultValue="code">
            <div className="flex items-center justify-between">
              <TabsList>
                <TabsTrigger value="code">Code</TabsTrigger>
                <TabsTrigger value="explanation">Explanation</TabsTrigger>
                {parentCode && (
                  <TabsTrigger value="diff">Diff</TabsTrigger>
                )}
              </TabsList>
            </div>
            <TabsContent value="code" className="mt-2">
              <pre className="max-h-[350px] overflow-auto rounded border bg-muted p-3 text-xs">
                <code>{code}</code>
              </pre>
            </TabsContent>
            <TabsContent value="explanation" className="mt-2">
              <div className="max-h-[350px] overflow-auto rounded border p-3 text-sm prose prose-sm dark:prose-invert">
                {explanation || "No explanation available."}
              </div>
            </TabsContent>
            {parentCode && (
              <TabsContent value="diff" className="mt-2">
                <DiffViewer oldCode={parentCode} newCode={code} />
              </TabsContent>
            )}
          </Tabs>
        </div>
      </div>

      {/* Bottom: Refine input */}
      <div className="flex gap-2">
        <Textarea
          placeholder="Describe modifications... e.g. 'Make the title bigger' or 'Add a legend'"
          value={modification}
          onChange={(e) => setModification(e.target.value)}
          className="min-h-[60px] flex-1 resize-none"
          rows={2}
        />
        <Button
          onClick={handleRefine}
          disabled={isRefining || !modification.trim()}
          className="self-end"
        >
          <RefreshCw
            className={`mr-2 h-4 w-4 ${isRefining ? "animate-spin" : ""}`}
          />
          {isRefining ? "Refining..." : "Refine"}
        </Button>
      </div>
    </div>
  );
}
