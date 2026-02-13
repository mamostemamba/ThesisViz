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
import { Input } from "@/components/ui/input";
import { useSettingsStore } from "@/stores/useSettingsStore";
import { useGenerateStore } from "@/stores/useGenerateStore";
import { useSavedSchemesStore, type SavedScheme } from "@/stores/useSavedSchemesStore";
import { extractColorsFromImage, deriveColorPair, hexToRgb } from "@/lib/color-extract";
import { Upload, Loader2, Trash2, Pencil, Check, X, Save, Plus, Minus } from "lucide-react";
import type { ColorPair, CustomColors } from "@/types/api";

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

/** Compact dot preview for saved scheme rows (up to 4 dots). */
function MiniPreview({ pairs }: { pairs: ColorPair[] }) {
  return (
    <div className="flex gap-1 shrink-0">
      {pairs.slice(0, 4).map((p, i) => (
        <span
          key={i}
          className="inline-block h-3 w-3 rounded-full border"
          style={{ backgroundColor: p.fill, borderColor: p.line }}
        />
      ))}
    </div>
  );
}

/** A single row in the saved schemes list with rename/delete. */
function SavedSchemeRow({
  scheme,
  onLoad,
}: {
  scheme: SavedScheme;
  onLoad: (pairs: ColorPair[]) => void;
}) {
  const rename = useSavedSchemesStore((s) => s.rename);
  const remove = useSavedSchemesStore((s) => s.remove);
  const [editing, setEditing] = useState(false);
  const [draft, setDraft] = useState(scheme.name);

  const handleRename = () => {
    const trimmed = draft.trim();
    if (trimmed) {
      rename(scheme.id, trimmed);
    }
    setEditing(false);
  };

  return (
    <div className="flex items-center gap-1.5 py-1 group">
      <MiniPreview pairs={scheme.pairs} />

      {editing ? (
        <>
          <Input
            value={draft}
            onChange={(e) => setDraft(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === "Enter") handleRename();
              if (e.key === "Escape") setEditing(false);
            }}
            className="h-6 text-xs flex-1 min-w-0 px-1"
            autoFocus
          />
          <button
            onClick={handleRename}
            className="p-0.5 text-muted-foreground hover:text-green-600"
          >
            <Check className="h-3 w-3" />
          </button>
          <button
            onClick={() => setEditing(false)}
            className="p-0.5 text-muted-foreground hover:text-red-500"
          >
            <X className="h-3 w-3" />
          </button>
        </>
      ) : (
        <>
          <button
            onClick={() => onLoad(scheme.pairs)}
            className="flex-1 text-left text-xs truncate hover:underline min-w-0"
            title={scheme.name}
          >
            {scheme.name}
          </button>
          <button
            onClick={() => { setDraft(scheme.name); setEditing(true); }}
            className="p-0.5 text-muted-foreground hover:text-foreground opacity-0 group-hover:opacity-100 transition-opacity"
          >
            <Pencil className="h-3 w-3" />
          </button>
          <button
            onClick={() => remove(scheme.id)}
            className="p-0.5 text-muted-foreground hover:text-red-500 opacity-0 group-hover:opacity-100 transition-opacity"
          >
            <Trash2 className="h-3 w-3" />
          </button>
        </>
      )}
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
  const isGenerating = useGenerateStore((s) => s.isGenerating);
  const isCancelled = useGenerateStore((s) => s.isCancelled);
  const isAnalyzing = useGenerateStore((s) => s.isAnalyzing);
  const phase = useGenerateStore((s) => s.phase);
  const generateError = useGenerateStore((s) => s.generateError);

  const savedSchemes = useSavedSchemesStore((s) => s.schemes);
  const saveScheme = useSavedSchemesStore((s) => s.save);

  const [extracting, setExtracting] = useState(false);
  const [extractError, setExtractError] = useState<string | null>(null);
  const [pendingColors, setPendingColors] = useState<ColorPair[] | null>(null);
  const [saveName, setSaveName] = useState("");
  const fileInputRef = useRef<HTMLInputElement>(null);

  const MIN_COLORS = 4;
  const MAX_COLORS = 8;
  const MANUAL_DEFAULTS = ["#4682B4", "#3CB371", "#FFA500", "#DC143C", "#9370DB", "#888888"];
  const [manualHexes, setManualHexes] = useState<string[]>(MANUAL_DEFAULTS);

  const isValidHex = (hex: string) => /^#[0-9a-fA-F]{6}$/.test(hex);

  /** Reverse-derive approximate base hex from a line color (line ≈ base × 0.7). */
  const lineToBaseHex = (line: string): string => {
    const r = parseInt(line.slice(1, 3), 16);
    const g = parseInt(line.slice(3, 5), 16);
    const b = parseInt(line.slice(5, 7), 16);
    const clamp = (v: number) => Math.min(255, Math.round(v / 0.7));
    const toHex2 = (v: number) => v.toString(16).padStart(2, "0");
    return `#${toHex2(clamp(r))}${toHex2(clamp(g))}${toHex2(clamp(b))}`;
  };

  const updateManualHex = (index: number, value: string) => {
    setManualHexes((prev) => {
      const next = [...prev];
      next[index] = value;
      return next;
    });
  };

  const addManualHex = () => {
    if (manualHexes.length >= MAX_COLORS) return;
    setManualHexes((prev) => [...prev, "#888888"]);
  };

  const removeManualHex = (index: number) => {
    if (manualHexes.length <= MIN_COLORS) return;
    setManualHexes((prev) => prev.filter((_, i) => i !== index));
  };

  const buildManualPairs = (): ColorPair[] =>
    manualHexes.map((hex) =>
      deriveColorPair(hexToRgb(isValidHex(hex) ? hex : "#888888"))
    );

  const handleManualApply = () => {
    setCustomColors({ pairs: buildManualPairs() });
  };

  const handleManualSaveAndApply = () => {
    const pairs = buildManualPairs();
    const name = `手动配色 ${savedSchemes.length + 1}`;
    saveScheme(name, pairs);
    setCustomColors({ pairs });
  };

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
    e.target.value = "";

    setExtracting(true);
    setExtractError(null);
    setPendingColors(null);

    try {
      const pairs = await extractColorsFromImage(file);
      setPendingColors(pairs);
    } catch (err) {
      setExtractError(
        err instanceof Error ? err.message : "Failed to extract colors"
      );
    } finally {
      setExtracting(false);
    }
  };

  const applyColors = (pairs: ColorPair[]) => {
    setCustomColors({ pairs });
    setPendingColors(null);
    setSaveName("");
  };

  const handleSaveAndApply = () => {
    if (!pendingColors) return;
    const name = saveName.trim() || `配色 ${savedSchemes.length + 1}`;
    saveScheme(name, pendingColors);
    applyColors(pendingColors);
  };

  const handleLoadSaved = (pairs: ColorPair[]) => {
    // Load into manual hex inputs so user can preview/modify before applying
    setManualHexes(pairs.map((p) => lineToBaseHex(p.line)));
    if (colorScheme !== "custom") {
      setColorScheme("custom");
    }
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
              <SelectItem value="custom">自定义</SelectItem>
            </SelectContent>
          </Select>

          {/* Custom color extraction UI */}
          {colorScheme === "custom" && (
            <div className="mt-3 space-y-3">
              {/* A) Current custom colors */}
              {customColors && (
                <div className="space-y-1">
                  <span className="text-xs text-muted-foreground">当前自定义配色</span>
                  <ColorPreview colors={customColors} />
                </div>
              )}

              {/* B) Manual hex input */}
              <div className="space-y-2 rounded border p-2">
                <div className="flex items-center justify-between">
                  <span className="text-xs font-medium text-muted-foreground">
                    手动输入颜色（{manualHexes.length}）
                  </span>
                  <button
                    onClick={addManualHex}
                    disabled={manualHexes.length >= MAX_COLORS}
                    className="p-0.5 text-muted-foreground hover:text-foreground disabled:opacity-30"
                    title="添加颜色"
                  >
                    <Plus className="h-3.5 w-3.5" />
                  </button>
                </div>
                {manualHexes.map((hex, i) => (
                  <div key={i} className="flex items-center gap-1">
                    <span className="text-xs w-4 shrink-0 text-muted-foreground">
                      {i + 1}
                    </span>
                    <Input
                      value={hex}
                      onChange={(e) => updateManualHex(i, e.target.value)}
                      className="h-6 text-xs flex-1 px-1 font-mono min-w-0"
                      placeholder="#000000"
                    />
                    <input
                      type="color"
                      value={isValidHex(hex) ? hex : "#888888"}
                      onChange={(e) => updateManualHex(i, e.target.value)}
                      className="h-6 w-6 shrink-0 cursor-pointer rounded border-0 p-0"
                    />
                    <button
                      onClick={() => removeManualHex(i)}
                      disabled={manualHexes.length <= MIN_COLORS}
                      className="p-0.5 text-muted-foreground hover:text-red-500 disabled:opacity-30"
                      title="删除"
                    >
                      <Minus className="h-3 w-3" />
                    </button>
                  </div>
                ))}
                <div className="flex gap-1.5 pt-1">
                  {manualHexes.map((hex, i) => {
                    const pair = isValidHex(hex)
                      ? deriveColorPair(hexToRgb(hex))
                      : null;
                    return pair ? (
                      <ColorDot key={i} fill={pair.fill} line={pair.line} />
                    ) : (
                      <span
                        key={i}
                        className="inline-block h-4 w-4 rounded-full border-2 border-dashed border-muted-foreground/40"
                      />
                    );
                  })}
                </div>
                <div className="flex gap-2">
                  <Button
                    variant="outline"
                    size="sm"
                    className="flex-1"
                    onClick={handleManualApply}
                  >
                    应用
                  </Button>
                  <Button
                    size="sm"
                    className="flex-1"
                    onClick={handleManualSaveAndApply}
                  >
                    <Save className="mr-1 h-3 w-3" />
                    保存并应用
                  </Button>
                </div>
              </div>

              {/* C) Upload + Extract + Save */}
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

              {extractError && (
                <p className="text-xs text-red-500">{extractError}</p>
              )}

              {pendingColors && (
                <div className="space-y-2 rounded border p-2">
                  <span className="text-xs text-muted-foreground">提取的配色</span>
                  <ColorPreview colors={{ pairs: pendingColors }} />

                  <Input
                    placeholder="配色名称（可选）"
                    value={saveName}
                    onChange={(e) => setSaveName(e.target.value)}
                    className="h-7 text-xs"
                  />

                  <div className="flex gap-2">
                    <Button
                      variant="outline"
                      size="sm"
                      className="flex-1"
                      onClick={() => applyColors(pendingColors)}
                    >
                      仅应用
                    </Button>
                    <Button
                      size="sm"
                      className="flex-1"
                      onClick={handleSaveAndApply}
                    >
                      <Save className="mr-1 h-3 w-3" />
                      保存并应用
                    </Button>
                  </div>
                </div>
              )}

              {/* C) Saved schemes list */}
              {savedSchemes.length > 0 && (
                <div className="space-y-1">
                  <span className="text-xs text-muted-foreground">
                    已保存配色（{savedSchemes.length}）
                  </span>
                  <div className="max-h-48 overflow-y-auto space-y-0.5">
                    {savedSchemes.map((scheme) => (
                      <SavedSchemeRow
                        key={scheme.id}
                        scheme={scheme}
                        onLoad={handleLoadSaved}
                      />
                    ))}
                  </div>
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
          {generateError ? (
            <Badge variant="destructive">出错</Badge>
          ) : isCancelled && !isGenerating ? (
            <Badge variant="outline" className="border-orange-400 text-orange-600">已终止</Badge>
          ) : isGenerating ? (
            <Badge variant={phase === "done" ? "outline" : "default"} className="gap-1">
              {phase !== "done" && (
                <Loader2 className="h-3 w-3 animate-spin" />
              )}
              {
                {
                  generating: "代码生成中",
                  compiling: "编译渲染中",
                  reviewing: "视觉审查中",
                  rerolling: "重新生成中",
                  fixing: "润色修复中",
                  done: "生成完成",
                }[phase] || "生成中"
              }
            </Badge>
          ) : isRendering ? (
            <Badge variant="default" className="gap-1">
              <Loader2 className="h-3 w-3 animate-spin" />
              渲染中
            </Badge>
          ) : isAnalyzing ? (
            <Badge variant="default" className="gap-1">
              <Loader2 className="h-3 w-3 animate-spin" />
              智能分析中
            </Badge>
          ) : (
            <Badge variant="secondary">就绪</Badge>
          )}
        </div>
      </div>
    </aside>
  );
}
