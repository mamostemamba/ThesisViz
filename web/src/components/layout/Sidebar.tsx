"use client";

import { useRef, useState } from "react";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { useSettingsStore } from "@/stores/useSettingsStore";
import { useGenerateStore } from "@/stores/useGenerateStore";
import { extractColors } from "@/lib/api";
import { Upload, Loader2 } from "lucide-react";
import type { CustomColors } from "@/types/api";

/** A small color swatch showing a fill/line pair. */
function ColorDot({ fill, line }: { fill: string; line: string }) {
  return (
    <span
      className="inline-block h-4 w-4 rounded-full border-2"
      style={{ backgroundColor: fill, borderColor: line }}
    />
  );
}

/** Preview row for custom color pairs (4-8). */
function ColorPreview({ colors }: { colors: CustomColors }) {
  return (
    <div className="flex gap-1.5 flex-wrap">
      {colors.pairs.map((p, i) => (
        <ColorDot key={i} fill={p.fill} line={p.line} />
      ))}
    </div>
  );
}

export function Sidebar() {
  const format = useSettingsStore((s) => s.format);
  const setFormat = useSettingsStore((s) => s.setFormat);
  const language = useSettingsStore((s) => s.language);
  const setLanguage = useSettingsStore((s) => s.setLanguage);
  const colorScheme = useSettingsStore((s) => s.colorScheme);
  const setColorScheme = useSettingsStore((s) => s.setColorScheme);
  const customColors = useSettingsStore((s) => s.customColors);
  const setCustomColors = useSettingsStore((s) => s.setCustomColors);
  const model = useSettingsStore((s) => s.model);
  const setModel = useSettingsStore((s) => s.setModel);
  const isRendering = useGenerateStore((s) => s.isRendering);

  const [extracting, setExtracting] = useState(false);
  const [extractError, setExtractError] = useState<string | null>(null);
  const [pendingColors, setPendingColors] = useState<CustomColors | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const handleSchemeChange = (value: string) => {
    setColorScheme(value);
    if (value !== "custom") {
      setCustomColors(null);
      setPendingColors(null);
    }
  };

  const handleFileSelect = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;
    // Reset input so re-selecting the same file triggers onChange
    e.target.value = "";

    setExtracting(true);
    setExtractError(null);
    setPendingColors(null);

    try {
      const res = await extractColors(file);
      setPendingColors(res.colors);
    } catch (err) {
      setExtractError(
        err instanceof Error ? err.message : "Failed to extract colors"
      );
    } finally {
      setExtracting(false);
    }
  };

  const handleApplyColors = () => {
    if (!pendingColors) return;
    setCustomColors(pendingColors);
    setPendingColors(null);
  };

  return (
    <aside className="w-64 border-r bg-muted/30 p-4">
      <div className="space-y-6">
        <div>
          <label className="mb-2 block text-sm font-medium">输出格式</label>
          <Select
            value={format}
            onValueChange={(v) =>
              setFormat(v as "tikz" | "matplotlib" | "mermaid")
            }
          >
            <SelectTrigger>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="tikz">TikZ</SelectItem>
              <SelectItem value="matplotlib">Matplotlib</SelectItem>
              <SelectItem value="mermaid">Mermaid</SelectItem>
            </SelectContent>
          </Select>
        </div>

        <div>
          <label className="mb-2 block text-sm font-medium">语言</label>
          <Select
            value={language}
            onValueChange={(v) => setLanguage(v as "en" | "zh")}
          >
            <SelectTrigger>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="en">英文</SelectItem>
              <SelectItem value="zh">中文</SelectItem>
            </SelectContent>
          </Select>
        </div>

        <div>
          <label className="mb-2 block text-sm font-medium">配色方案</label>
          <Select value={colorScheme} onValueChange={handleSchemeChange}>
            <SelectTrigger>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="drawio">Draw.io 经典</SelectItem>
              <SelectItem value="professional_blue">专业蓝</SelectItem>
              <SelectItem value="bold_contrast">高对比</SelectItem>
              <SelectItem value="minimal_mono">极简黑白</SelectItem>
              <SelectItem value="modern_teal">现代青</SelectItem>
              <SelectItem value="soft_pastel">柔和粉彩</SelectItem>
              <SelectItem value="warm_earth">暖色大地</SelectItem>
              <SelectItem value="cyber_dark">深色科技</SelectItem>
              <SelectItem value="custom">自定义（从图片提取）</SelectItem>
            </SelectContent>
          </Select>

          {/* Custom color extraction UI */}
          {colorScheme === "custom" && (
            <div className="mt-3 space-y-2">
              {/* Show current custom colors if already applied */}
              {customColors && (
                <div className="space-y-1">
                  <span className="text-xs text-muted-foreground">当前自定义配色</span>
                  <ColorPreview colors={customColors} />
                </div>
              )}

              {/* Upload button */}
              <input
                ref={fileInputRef}
                type="file"
                accept="image/*"
                className="hidden"
                onChange={handleFileSelect}
              />
              <Button
                variant="outline"
                size="sm"
                className="w-full"
                onClick={() => fileInputRef.current?.click()}
                disabled={extracting}
              >
                {extracting ? (
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                ) : (
                  <Upload className="mr-2 h-4 w-4" />
                )}
                {extracting ? "提取中..." : "上传参考图"}
              </Button>

              {/* Error */}
              {extractError && (
                <p className="text-xs text-red-500">{extractError}</p>
              )}

              {/* Pending colors preview + apply */}
              {pendingColors && (
                <div className="space-y-2 rounded border p-2">
                  <span className="text-xs text-muted-foreground">提取的配色</span>
                  <ColorPreview colors={pendingColors} />
                  <Button
                    size="sm"
                    className="w-full"
                    onClick={handleApplyColors}
                  >
                    应用配色
                  </Button>
                </div>
              )}
            </div>
          )}
        </div>

        <div>
          <label className="mb-2 block text-sm font-medium">AI 模型</label>
          <Select value={model} onValueChange={setModel}>
            <SelectTrigger>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="gemini-3-pro-preview">Gemini 3 Pro (Preview)</SelectItem>
              <SelectItem value="gemini-3-flash-preview">Gemini 3 Flash (Preview)</SelectItem>
              <SelectItem value="gemini-2.5-pro">Gemini 2.5 Pro</SelectItem>
              <SelectItem value="gemini-2.5-flash">Gemini 2.5 Flash</SelectItem>
              <SelectItem value="gemini-2.5-flash-lite">Gemini 2.5 Flash Lite</SelectItem>
            </SelectContent>
          </Select>
        </div>

        <div>
          <label className="mb-2 block text-sm font-medium">状态</label>
          <Badge variant={isRendering ? "default" : "secondary"}>
            {isRendering ? "渲染中..." : "就绪"}
          </Badge>
        </div>
      </div>
    </aside>
  );
}
