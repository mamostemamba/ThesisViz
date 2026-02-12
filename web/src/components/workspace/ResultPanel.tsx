"use client";

import { useState, useCallback } from "react";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Copy, Check, RefreshCw } from "lucide-react";

interface ResultPanelProps {
  code: string;
  format: string;
  explanation: string;
  imageUrl: string;
  reviewPassed: boolean;
  reviewRounds: number;
  onRefine: (modification: string) => void;
  isRefining: boolean;
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
}: ResultPanelProps) {
  const [modification, setModification] = useState("");
  const [copied, setCopied] = useState(false);

  const handleCopy = useCallback(async () => {
    await navigator.clipboard.writeText(code);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
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
          {imageUrl ? (
            <a
              href={imageUrl}
              target="_blank"
              rel="noopener noreferrer"
              className="block"
            >
              <img
                src={imageUrl}
                alt="Generated figure"
                className="max-h-[400px] w-full rounded border object-contain"
              />
            </a>
          ) : (
            <div className="flex h-[300px] items-center justify-center rounded border border-dashed text-sm text-muted-foreground">
              {format === "mermaid"
                ? "Mermaid diagram (render in browser)"
                : "No image available"}
            </div>
          )}
        </div>

        {/* Right: Code / Explanation tabs */}
        <div className="space-y-2">
          <Tabs defaultValue="code">
            <div className="flex items-center justify-between">
              <TabsList>
                <TabsTrigger value="code">Code</TabsTrigger>
                <TabsTrigger value="explanation">Explanation</TabsTrigger>
              </TabsList>
              <Button variant="ghost" size="sm" onClick={handleCopy}>
                {copied ? (
                  <Check className="mr-1 h-3 w-3" />
                ) : (
                  <Copy className="mr-1 h-3 w-3" />
                )}
                {copied ? "Copied" : "Copy"}
              </Button>
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
          <RefreshCw className={`mr-2 h-4 w-4 ${isRefining ? "animate-spin" : ""}`} />
          {isRefining ? "Refining..." : "Refine"}
        </Button>
      </div>
    </div>
  );
}
