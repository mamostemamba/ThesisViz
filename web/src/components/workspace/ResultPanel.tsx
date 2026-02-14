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
  ZoomIn,
  MessageSquareText,
  AlertCircle,
} from "lucide-react";
import { ImageLightbox } from "./ImageLightbox";
import { MermaidRenderer } from "./MermaidRenderer";
import { DiffViewer } from "./DiffViewer";
import { TikZEditor, type HighlightState } from "./TikZEditor";
import { useSettingsStore } from "@/stores/useSettingsStore";

export interface ImageSnapshot {
  round: number;
  imageUrl: string;
  label: string;
}

interface ResultPanelProps {
  code: string;
  format: string;
  explanation?: string;
  imageUrl: string;
  reviewPassed: boolean;
  reviewRounds: number;
  reviewCritique?: string;
  reviewIssues?: string[];
  onRefine: (modification: string) => void;
  isRefining: boolean;
  parentCode?: string;
  imageSnapshots?: ImageSnapshot[];
  fullTex?: string;
}

export function ResultPanel({
  code,
  format,
  imageUrl,
  reviewPassed,
  reviewRounds,
  onRefine,
  isRefining,
  parentCode,
  imageSnapshots,
  reviewCritique,
  reviewIssues,
  fullTex,
}: ResultPanelProps) {
  const colorScheme = useSettingsStore((s) => s.colorScheme);
  const customColors = useSettingsStore((s) => s.customColors);
  const language = useSettingsStore((s) => s.language);
  const diagramStyle = useSettingsStore((s) => s.diagramStyle);
  const [modification, setModification] = useState("");
  const [copied, setCopied] = useState<string | null>(null);
  const [lightboxSrc, setLightboxSrc] = useState<string | null>(null);
  const mermaidContainerRef = useRef<HTMLDivElement>(null);

  // Editable state for TikZ fine-tuning
  const [editedCode, setEditedCode] = useState<string | null>(null);
  const [editedImageUrl, setEditedImageUrl] = useState<string | null>(null);

  // Highlight state for element location
  const [highlightState, setHighlightState] = useState<HighlightState>({
    loading: false,
    imageUrl: null,
  });

  // Use edited values if available, otherwise fall back to props
  const currentImageUrl = editedImageUrl || imageUrl;
  const currentCode = editedCode || code;

  // Display image: highlight overrides when available
  const displayImageUrl = highlightState.imageUrl || currentImageUrl;

  const handleCopy = useCallback(
    async (label: string, text: string) => {
      await navigator.clipboard.writeText(text);
      setCopied(label);
      setTimeout(() => setCopied(null), 2000);
    },
    []
  );

  // Code panel and copy button use fullTex when available (Overleaf-ready)
  const displayCode = fullTex || currentCode;

  const handleCopyCode = useCallback(() => {
    handleCopy("code", displayCode);
  }, [displayCode, handleCopy]);

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
    } else if (currentImageUrl) {
      // For TikZ/Matplotlib, download the server image
      const response = await fetch(currentImageUrl);
      const blob = await response.blob();
      const url = URL.createObjectURL(blob);
      const link = document.createElement("a");
      link.download = "figure.png";
      link.href = url;
      link.click();
      URL.revokeObjectURL(url);
    }
  }, [format, currentImageUrl]);

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
            <h3 className="text-sm font-semibold">结果</h3>
            <div className="flex items-center gap-2 text-xs text-muted-foreground">
              {reviewPassed ? (
                <span className="text-green-600">审查通过</span>
              ) : (
                <span className="text-yellow-600">
                  审查：{reviewRounds} 轮
                </span>
              )}
            </div>
          </div>
          {format === "mermaid" && code ? (
            <div ref={mermaidContainerRef}>
              <MermaidRenderer code={code} colorScheme={colorScheme} />
            </div>
          ) : displayImageUrl ? (
            <div className="relative group">
              {/* eslint-disable-next-line @next/next/no-img-element */}
              <img
                src={displayImageUrl}
                alt="Generated figure"
                className={`max-h-[400px] w-full rounded border object-contain transition-all duration-300 ${
                  highlightState.loading
                    ? "border-red-300 opacity-70"
                    : highlightState.imageUrl
                      ? "border-red-400 border-2"
                      : ""
                }`}
              />
              {highlightState.loading && (
                <div className="absolute inset-0 flex items-center justify-center pointer-events-none">
                  <span className="animate-pulse rounded-full bg-red-500/20 px-3 py-1 text-xs font-medium text-red-600">
                    定位中...
                  </span>
                </div>
              )}
              <button
                onClick={() => setLightboxSrc(displayImageUrl)}
                className="absolute top-2 right-2 rounded-md bg-black/50 p-1.5 text-white opacity-0 transition-opacity group-hover:opacity-100 hover:bg-black/70"
                title="放大查看"
              >
                <ZoomIn className="h-4 w-4" />
              </button>
            </div>
          ) : (
            <div className="flex h-[300px] items-center justify-center rounded border border-dashed text-sm text-muted-foreground">
              暂无图片
            </div>
          )}

          {/* Version history: show all round images */}
          {imageSnapshots && imageSnapshots.length > 1 && (
            <details className="rounded border bg-muted/30">
              <summary className="cursor-pointer px-3 py-2 text-xs font-medium text-muted-foreground hover:text-foreground">
                优化历程 ({imageSnapshots.length} 版)
              </summary>
              <div className="grid gap-2 p-2 sm:grid-cols-2 border-t">
                {imageSnapshots.map((snap, i) => (
                  <div key={snap.imageUrl} className="space-y-1">
                    <p className="text-xs text-muted-foreground">
                      {i + 1}. {snap.label}
                      {i === imageSnapshots.length - 1 && (
                        <span className="ml-1 text-green-600 font-medium">（最终版）</span>
                      )}
                    </p>
                    <div className="relative group">
                      <img
                        src={snap.imageUrl}
                        alt={snap.label}
                        className="max-h-[200px] rounded border object-contain w-full"
                      />
                      <button
                        onClick={() => setLightboxSrc(snap.imageUrl)}
                        className="absolute top-1 right-1 rounded bg-black/50 p-1 text-white opacity-0 transition-opacity group-hover:opacity-100 hover:bg-black/70"
                        title="放大查看"
                      >
                        <ZoomIn className="h-3 w-3" />
                      </button>
                    </div>
                  </div>
                ))}
              </div>
            </details>
          )}

          {/* Review issues */}
          {reviewIssues && reviewIssues.length > 0 && (
            <div className="rounded border border-yellow-200 bg-yellow-50 p-2.5 text-xs dark:border-yellow-800 dark:bg-yellow-950">
              <div className="flex items-center gap-1 font-medium text-yellow-700 dark:text-yellow-300 mb-1">
                <AlertCircle className="h-3 w-3" />
                审查发现的问题（{reviewIssues.length} 项）
              </div>
              <ul className="list-disc list-inside space-y-0.5 text-yellow-600 dark:text-yellow-400">
                {reviewIssues.map((issue, i) => (
                  <li key={i}>{issue}</li>
                ))}
              </ul>
            </div>
          )}

          {/* AI Critique */}
          {reviewCritique && (
            <div className="rounded border border-blue-200 bg-blue-50 p-2.5 text-xs dark:border-blue-800 dark:bg-blue-950">
              <div className="flex items-center gap-1 font-medium text-blue-700 dark:text-blue-300 mb-1">
                <MessageSquareText className="h-3 w-3" />
                AI 点评
              </div>
              <p className="text-blue-600 dark:text-blue-400">{reviewCritique}</p>
            </div>
          )}

          {/* Export buttons */}
          <div className="flex flex-wrap gap-2">
            {(currentImageUrl || format === "mermaid") && (
              <Button variant="outline" size="sm" onClick={handleDownloadPNG}>
                <Download className="mr-1 h-3 w-3" />
                下载 PNG
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
                  下载 SVG
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
                <TabsTrigger value="code">代码</TabsTrigger>
                {parentCode && (
                  <TabsTrigger value="diff">差异对比</TabsTrigger>
                )}
                {format === "tikz" && (
                  <TabsTrigger value="editor">微调</TabsTrigger>
                )}
              </TabsList>
            </div>
            <TabsContent value="code" className="mt-2">
              <div className="relative group">
                <pre className="max-h-[350px] overflow-auto rounded border bg-muted p-3 text-xs">
                  <code>{displayCode}</code>
                </pre>
                <button
                  onClick={handleCopyCode}
                  className="absolute top-2 right-2 rounded bg-background/80 border p-1.5 text-muted-foreground opacity-0 transition-opacity group-hover:opacity-100 hover:text-foreground"
                  title="复制代码"
                >
                  {copied === "code" ? <Check className="h-3.5 w-3.5 text-green-500" /> : <Copy className="h-3.5 w-3.5" />}
                </button>
              </div>
            </TabsContent>
            {parentCode && (
              <TabsContent value="diff" className="mt-2">
                <DiffViewer oldCode={parentCode} newCode={code} />
              </TabsContent>
            )}
            {format === "tikz" && (
              <TabsContent value="editor" className="mt-2">
                <TikZEditor
                  code={currentCode}
                  imageUrl={currentImageUrl}
                  onCodeChange={setEditedCode}
                  onImageChange={setEditedImageUrl}
                  onHighlightChange={setHighlightState}
                  format={format}
                  language={language}
                  colorScheme={colorScheme}
                  customColors={colorScheme === "custom" && customColors ? customColors : undefined}
                  diagramStyle={diagramStyle}
                />
              </TabsContent>
            )}
          </Tabs>
        </div>
      </div>

      {/* Bottom: Refine input */}
      <div className="flex gap-2">
        <Textarea
          placeholder="描述修改内容，例如「把标题放大」或「添加图例」"
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
          {isRefining ? "优化中..." : "优化"}
        </Button>
      </div>

      {lightboxSrc && (
        <ImageLightbox
          src={lightboxSrc}
          alt="放大查看"
          onClose={() => setLightboxSrc(null)}
        />
      )}
    </div>
  );
}
