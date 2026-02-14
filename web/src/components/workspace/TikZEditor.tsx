"use client";

import { useState, useCallback, useEffect, useMemo, useRef } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import {
  Circle,
  ArrowRight,
  Crosshair,
  FileText,
  RefreshCw,
  Loader2,
  Undo2,
  Search,
} from "lucide-react";
import { parseTikZ } from "@/lib/tikz/parser";
import { compileTikZ, compileTikZWithHighlight, compileOptions } from "@/lib/tikz/compiler";
import { renderCode } from "@/lib/api";
import type {
  TikZDocument,
  TikZElement,
  TikZNode,
  TikZDraw,
  TikZOption,
} from "@/lib/tikz/types";

export interface HighlightState {
  loading: boolean;
  imageUrl: string | null;
}

interface TikZEditorProps {
  code: string;
  imageUrl: string;
  onCodeChange: (code: string) => void;
  onImageChange: (url: string) => void;
  onHighlightChange?: (state: HighlightState) => void;
  format: string;
  language: string;
  colorScheme: string;
  customColors?: import("@/types/api").CustomColors;
  diagramStyle: string;
}

/** Get a human-readable label for an element. */
function getLabel(el: TikZElement, nodeMap?: Map<string, string>): string {
  switch (el.type) {
    case "node":
      return el.content
        ? el.content.slice(0, 30) + (el.content.length > 30 ? "..." : "")
        : el.name || "(未命名节点)";
    case "draw":
      return getDrawLabel(el, undefined, nodeMap);
    case "coordinate":
      return el.name || "(未命名坐标)";
    case "raw":
      return el.content.slice(0, 40) + (el.content.length > 40 ? "..." : "");
  }
}

/**
 * Build a node name → content map from ALL sources:
 *   1. Parsed \node elements (name → content)
 *   2. \matrix cells: auto-named matrixName-row-col → cell label
 *
 * The V2 renderer emits matrices like:
 *   \matrix (dec) [matrix of nodes, ...] {
 *     |[fill=primaryFill, draw=primaryLine]| 加法与归一化层 \\
 *     |[fill=secondaryFill, draw=secondaryLine]| 前馈网络 \\
 *   };
 * Cells are auto-named dec-1-1, dec-2-1, etc.
 */
function buildNodeContentMap(doc: TikZDocument): Map<string, string> {
  const map = new Map<string, string>();

  for (const el of doc.elements) {
    // 1. Parsed nodes
    if (el.type === "node" && el.name) {
      map.set(el.name, el.content || el.name);
      continue;
    }

    // 2. Matrices inside raw elements
    if (el.type !== "raw") continue;
    const text = el.content;
    const mxMatch = text.match(/\\matrix\s*\(([^)]+)\)/);
    if (!mxMatch) continue;

    const mxName = mxMatch[1].trim();

    // Skip past [options] to find the matrix body { ... }
    // Options may contain nested braces like nodes={matrix_node}
    let searchStart = mxMatch.index! + mxMatch[0].length;
    // Skip whitespace
    while (searchStart < text.length && /\s/.test(text[searchStart])) searchStart++;
    // Skip [options] block (may contain nested {})
    if (text[searchStart] === "[") {
      let bd = 1;
      searchStart++;
      while (searchStart < text.length && bd > 0) {
        if (text[searchStart] === "[") bd++;
        else if (text[searchStart] === "]") bd--;
        searchStart++;
      }
    }

    // Now find the body { ... }
    let braceStart = -1;
    let depth = 0;
    for (let i = searchStart; i < text.length; i++) {
      if (text[i] === "{") {
        if (depth === 0) braceStart = i;
        depth++;
      } else if (text[i] === "}") {
        depth--;
        if (depth === 0 && braceStart !== -1) {
          const body = text.slice(braceStart + 1, i);
          // Split rows by \\  (not inside braces)
          const rows = body.split(/\\\\\s*/);
          let row = 0;
          for (const rowStr of rows) {
            // Strip comment lines before processing
            const trimmed = rowStr
              .split("\n")
              .filter((l) => !l.trim().startsWith("%"))
              .join("\n")
              .trim();
            if (!trimmed) continue;
            row++;
            // Split cols by &
            const cols = trimmed.split(/\s*&\s*/);
            let col = 0;
            for (const cell of cols) {
              const c = cell.trim();
              if (!c) continue;
              col++;
              // Extract label: |[opts]| label  or  {label}
              const pipeMatch = c.match(/\|\[[^\]]*\]\|\s*(.+)/);
              const label = pipeMatch ? pipeMatch[1].trim() : "";
              if (label && label !== "" && !label.startsWith("%")) {
                // Strip surrounding braces if present
                const clean = label.replace(/^\{|\}$/g, "").trim();
                if (clean) {
                  map.set(`${mxName}-${row}-${col}`, clean);
                }
              }
            }
          }
          break;
        }
      }
    }
  }

  return map;
}

/**
 * Build a human-readable label for a draw element by resolving
 * the referenced node names to their text content.
 * e.g. "(input.east) -- (proof.west)" → "输入嵌入 → 证明系统"
 */
function getDrawLabel(draw: TikZDraw, doc?: TikZDocument, nodeMap?: Map<string, string>): string {
  const nodeContent = nodeMap ?? new Map<string, string>();

  // Extract (nodename) or (nodename.anchor) references from the path
  const refs: string[] = [];
  for (const m of draw.path.matchAll(/\(([^)]+)\)/g)) {
    const inner = m[1].trim();
    // Skip coordinates like (0,0), calc expressions like ($(...)$)
    if (inner.includes(",") || inner.startsWith("$")) continue;
    // Strip anchor: "node.east" → "node", "node.north east" → "node"
    const name = inner.split(".")[0].trim();
    if (name && !refs.includes(name)) refs.push(name);
  }

  // Resolve to content labels
  const labels = refs.map((name) => {
    const content = nodeContent.get(name);
    if (content) {
      return content.length > 12 ? content.slice(0, 12) + "…" : content;
    }
    return name;
  });

  if (labels.length >= 2) {
    return `${labels[0]} → ${labels[labels.length - 1]}`;
  }
  if (labels.length === 1) {
    return labels[0];
  }
  // Fallback: raw path
  return draw.path.slice(0, 40) + (draw.path.length > 40 ? "..." : "");
}

function ElementIcon({ type }: { type: TikZElement["type"] }) {
  switch (type) {
    case "node":
      return <Circle className="h-3 w-3 text-blue-500" />;
    case "draw":
      return <ArrowRight className="h-3 w-3 text-green-500" />;
    case "coordinate":
      return <Crosshair className="h-3 w-3 text-orange-500" />;
    case "raw":
      return <FileText className="h-3 w-3 text-gray-400" />;
  }
}

const TYPE_LABELS: Record<TikZElement["type"], string> = {
  node: "节点",
  draw: "连线",
  coordinate: "坐标",
  raw: "原始",
};

function getOpt(opts: TikZOption[], key: string): string {
  return opts.find((o) => o.key === key)?.value ?? "";
}

function setOpt(opts: TikZOption[], key: string, value: string): TikZOption[] {
  const idx = opts.findIndex((o) => o.key === key);
  if (value === "") {
    if (idx !== -1) return opts.filter((_, i) => i !== idx);
    return opts;
  }
  if (idx !== -1) {
    return opts.map((o, i) => (i === idx ? { ...o, value } : o));
  }
  return [...opts, { key, value }];
}

// ---------------------------------------------------------------------------
// Property panels
// ---------------------------------------------------------------------------

function NodePropertyPanel({
  node,
  onChange,
}: {
  node: TikZNode;
  onChange: (updated: TikZNode) => void;
}) {
  const handleContentChange = (value: string) => {
    onChange({ ...node, content: value });
  };

  const handlePositionChange = (value: string) => {
    onChange({ ...node, position: value });
  };

  const handleOptChange = (key: string, value: string) => {
    onChange({ ...node, options: setOpt(node.options, key, value) });
  };

  return (
    <div className="space-y-3">
      <h4 className="text-xs font-semibold text-muted-foreground tracking-wide">
        节点属性
      </h4>

      {node.name && (
        <div>
          <label className="text-xs text-muted-foreground">名称</label>
          <Input value={node.name} disabled className="h-7 text-xs bg-muted" />
        </div>
      )}

      <div>
        <label className="text-xs text-muted-foreground">文本内容</label>
        <Textarea
          value={node.content}
          onChange={(e) => handleContentChange(e.target.value)}
          className="min-h-[40px] text-xs"
          rows={2}
        />
      </div>

      {node.position && (
        <div>
          <label className="text-xs text-muted-foreground">位置</label>
          <Input
            value={node.position}
            onChange={(e) => handlePositionChange(e.target.value)}
            className="h-7 text-xs"
          />
        </div>
      )}

      <div className="grid grid-cols-2 gap-2">
        <div>
          <label className="text-xs text-muted-foreground">填充色</label>
          <Input
            value={getOpt(node.options, "fill")}
            onChange={(e) => handleOptChange("fill", e.target.value)}
            className="h-7 text-xs"
            placeholder="无"
          />
        </div>
        <div>
          <label className="text-xs text-muted-foreground">边框色</label>
          <Input
            value={getOpt(node.options, "draw")}
            onChange={(e) => handleOptChange("draw", e.target.value)}
            className="h-7 text-xs"
            placeholder="无"
          />
        </div>
      </div>

      <div className="grid grid-cols-2 gap-2">
        <div>
          <label className="text-xs text-muted-foreground">最小宽度</label>
          <Input
            value={getOpt(node.options, "minimum width")}
            onChange={(e) => handleOptChange("minimum width", e.target.value)}
            className="h-7 text-xs"
            placeholder="如 3cm"
          />
        </div>
        <div>
          <label className="text-xs text-muted-foreground">最小高度</label>
          <Input
            value={getOpt(node.options, "minimum height")}
            onChange={(e) => handleOptChange("minimum height", e.target.value)}
            className="h-7 text-xs"
            placeholder="如 1cm"
          />
        </div>
      </div>

      <div>
        <label className="text-xs text-muted-foreground">字体</label>
        <Input
          value={getOpt(node.options, "font")}
          onChange={(e) => handleOptChange("font", e.target.value)}
          className="h-7 text-xs"
          placeholder="如 \\footnotesize"
        />
      </div>

      <details className="text-xs">
        <summary className="cursor-pointer text-muted-foreground hover:text-foreground">
          全部选项（{node.options.length} 项）
        </summary>
        <pre className="mt-1 rounded bg-muted p-2 text-[10px] overflow-auto max-h-[120px]">
          {compileOptions(node.options) || "（无）"}
        </pre>
      </details>
    </div>
  );
}

function DrawPropertyPanel({
  draw,
  onChange,
}: {
  draw: TikZDraw;
  onChange: (updated: TikZDraw) => void;
}) {
  const handleOptChange = (key: string, value: string) => {
    onChange({ ...draw, options: setOpt(draw.options, key, value) });
  };

  const handlePathChange = (value: string) => {
    onChange({ ...draw, path: value });
  };

  return (
    <div className="space-y-3">
      <h4 className="text-xs font-semibold text-muted-foreground tracking-wide">
        连线属性
      </h4>

      <div>
        <label className="text-xs text-muted-foreground">命令</label>
        <Input
          value={`\\${draw.command}`}
          disabled
          className="h-7 text-xs bg-muted"
        />
      </div>

      <div>
        <label className="text-xs text-muted-foreground">路径</label>
        <Textarea
          value={draw.path}
          onChange={(e) => handlePathChange(e.target.value)}
          className="min-h-[40px] text-xs"
          rows={2}
        />
      </div>

      <div className="grid grid-cols-2 gap-2">
        <div>
          <label className="text-xs text-muted-foreground">颜色</label>
          <Input
            value={getOpt(draw.options, "color")}
            onChange={(e) => handleOptChange("color", e.target.value)}
            className="h-7 text-xs"
            placeholder="black"
          />
        </div>
        <div>
          <label className="text-xs text-muted-foreground">线宽</label>
          <Input
            value={getOpt(draw.options, "line width")}
            onChange={(e) => handleOptChange("line width", e.target.value)}
            className="h-7 text-xs"
            placeholder="如 0.5pt"
          />
        </div>
      </div>

      <details className="text-xs">
        <summary className="cursor-pointer text-muted-foreground hover:text-foreground">
          全部选项（{draw.options.length} 项）
        </summary>
        <pre className="mt-1 rounded bg-muted p-2 text-[10px] overflow-auto max-h-[120px]">
          {compileOptions(draw.options) || "（无）"}
        </pre>
      </details>
    </div>
  );
}

function RawPropertyPanel({
  el,
  onChange,
}: {
  el: import("@/lib/tikz/types").TikZRaw;
  onChange: (updated: import("@/lib/tikz/types").TikZRaw) => void;
}) {
  return (
    <div className="space-y-3">
      <h4 className="text-xs font-semibold text-muted-foreground tracking-wide">
        原始代码
      </h4>
      <Textarea
        value={el.content}
        onChange={(e) => onChange({ ...el, content: e.target.value })}
        className="min-h-[80px] text-xs font-mono"
        rows={4}
      />
    </div>
  );
}

// ---------------------------------------------------------------------------
// Main TikZEditor component
// ---------------------------------------------------------------------------

export function TikZEditor({
  code,
  imageUrl,
  onCodeChange,
  onImageChange,
  onHighlightChange,
  format,
  language,
  colorScheme,
  customColors,
  diagramStyle,
}: TikZEditorProps) {
  const [doc, setDoc] = useState<TikZDocument | null>(null);
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [isRendering, setIsRendering] = useState(false);
  const [renderError, setRenderError] = useState<string | null>(null);
  const [originalCode] = useState(code);
  const [showRaw, setShowRaw] = useState(false);
  const [typeFilter, setTypeFilter] = useState<"all" | "node" | "draw" | "coordinate">("node");
  const [searchQuery, setSearchQuery] = useState("");

  // Refs for highlight rendering (avoid stale closures)
  const codeRef = useRef(code);
  codeRef.current = code;
  const docRef = useRef(doc);
  docRef.current = doc;
  const highlightSeqRef = useRef(0);
  const renderParamsRef = useRef({ format, language, colorScheme, customColors, diagramStyle });
  renderParamsRef.current = { format, language, colorScheme, customColors, diagramStyle };

  // Parse on mount
  useEffect(() => {
    const parsed = parseTikZ(code);
    setDoc(parsed);
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  const stats = useMemo(() => {
    if (!doc) return { nodes: 0, draws: 0, coords: 0, raw: 0 };
    return {
      nodes: doc.elements.filter((e) => e.type === "node").length,
      draws: doc.elements.filter((e) => e.type === "draw").length,
      coords: doc.elements.filter((e) => e.type === "coordinate").length,
      raw: doc.elements.filter((e) => e.type === "raw").length,
    };
  }, [doc]);

  // Unified node name → content map (parsed nodes + matrix cells)
  const nodeContentMap = useMemo(
    () => (doc ? buildNodeContentMap(doc) : new Map<string, string>()),
    [doc]
  );

  const filteredElements = useMemo(() => {
    if (!doc) return [];
    const q = searchQuery.toLowerCase();
    return doc.elements.filter((el) => {
      // Always hide raw unless showRaw is on
      if (el.type === "raw") return showRaw && !!el.content.trim();
      // Type filter
      if (typeFilter !== "all" && el.type !== typeFilter) return false;
      // Search filter
      if (q) {
        const text =
          el.type === "node"
            ? `${el.name} ${el.content}`
            : el.type === "draw"
              ? `${el.path} ${getDrawLabel(el, undefined, nodeContentMap)}`
              : el.type === "coordinate"
                ? el.name
                : "";
        if (!text.toLowerCase().includes(q)) return false;
      }
      return true;
    });
  }, [doc, typeFilter, searchQuery, showRaw]);

  const selectedElement = useMemo(
    () => doc?.elements.find((e) => e.id === selectedId) ?? null,
    [doc, selectedId]
  );

  const handleElementChange = useCallback((updated: TikZElement) => {
    setDoc((prev) => {
      if (!prev) return prev;
      return {
        ...prev,
        elements: prev.elements.map((el) =>
          el.id === updated.id ? updated : el
        ),
      };
    });
  }, []);

  const handleRender = useCallback(async () => {
    if (!doc) return;
    const newCode = compileTikZ(doc);
    setIsRendering(true);
    setRenderError(null);

    try {
      const res = await renderCode({
        code: newCode,
        format,
        language,
        color_scheme: colorScheme,
        custom_colors:
          colorScheme === "custom" && customColors ? customColors : undefined,
        style: diagramStyle,
      });

      if (res.status === "ok" && res.image_url) {
        onCodeChange(newCode);
        onImageChange(res.image_url);
        // Clear stale highlight since the base image changed
        onHighlightChange?.({ loading: false, imageUrl: null });
      } else {
        setRenderError(res.error || "渲染失败");
      }
    } catch (err) {
      setRenderError(
        err instanceof Error ? err.message : "渲染请求失败"
      );
    } finally {
      setIsRendering(false);
    }
  }, [
    doc,
    format,
    language,
    colorScheme,
    customColors,
    diagramStyle,
    onCodeChange,
    onImageChange,
  ]);

  const handleReset = useCallback(() => {
    const parsed = parseTikZ(originalCode);
    setDoc(parsed);
    setSelectedId(null);
  }, [originalCode]);

  // Highlight element on image when selection changes
  useEffect(() => {
    if (!onHighlightChange) return;

    // Increment sequence to invalidate in-flight renders
    const seq = ++highlightSeqRef.current;

    if (!selectedId || !docRef.current) {
      onHighlightChange({ loading: false, imageUrl: null });
      return;
    }

    // Raw elements can't be highlighted
    const target = docRef.current.elements.find((e) => e.id === selectedId);
    if (!target || target.type === "raw") {
      onHighlightChange({ loading: false, imageUrl: null });
      return;
    }

    // Immediate loading feedback
    onHighlightChange({ loading: true, imageUrl: null });

    // Debounced render
    const timer = setTimeout(async () => {
      const currentDoc = docRef.current;
      const p = renderParamsRef.current;
      if (!currentDoc) return;

      try {
        const hlCode = compileTikZWithHighlight(codeRef.current, currentDoc, selectedId);
        const res = await renderCode({
          code: hlCode,
          format: p.format,
          language: p.language,
          color_scheme: p.colorScheme,
          custom_colors:
            p.colorScheme === "custom" && p.customColors
              ? p.customColors
              : undefined,
          style: p.diagramStyle,
        });

        // Stale? Ignore.
        if (highlightSeqRef.current !== seq) return;

        if (res.status === "ok" && res.image_url) {
          onHighlightChange({ loading: false, imageUrl: res.image_url });
        } else {
          onHighlightChange({ loading: false, imageUrl: null });
        }
      } catch {
        if (highlightSeqRef.current === seq) {
          onHighlightChange({ loading: false, imageUrl: null });
        }
      }
    }, 200);

    return () => clearTimeout(timer);
  }, [selectedId, onHighlightChange]); // eslint-disable-line react-hooks/exhaustive-deps

  // Clear highlight on unmount
  useEffect(() => {
    return () => {
      onHighlightChange?.({ loading: false, imageUrl: null });
    };
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  if (!doc) return null;

  return (
    <div className="space-y-3">
      {/* Type filter + search + actions */}
      <div className="space-y-2">
        {/* Filter tabs */}
        <div className="flex items-center gap-1 flex-wrap">
          {([
            { key: "node" as const, icon: Circle, label: "节点", count: stats.nodes, color: "text-blue-500" },
            { key: "draw" as const, icon: ArrowRight, label: "连线", count: stats.draws, color: "text-green-500" },
            { key: "coordinate" as const, icon: Crosshair, label: "坐标", count: stats.coords, color: "text-orange-500" },
            { key: "all" as const, icon: FileText, label: "全部", count: stats.nodes + stats.draws + stats.coords, color: "text-muted-foreground" },
          ]).map(({ key, icon: Icon, label, count, color }) => (
            <button
              key={key}
              onClick={() => setTypeFilter(key)}
              className={`flex items-center gap-1 rounded-md px-2 py-1 text-xs transition-colors ${
                typeFilter === key
                  ? "bg-primary/10 text-primary font-medium"
                  : "text-muted-foreground hover:bg-muted"
              }`}
            >
              <Icon className={`h-3 w-3 ${typeFilter === key ? "text-primary" : color}`} />
              {label}
              <span className="text-[10px] opacity-60">{count}</span>
            </button>
          ))}
          {stats.raw > 0 && (
            <button
              onClick={() => setShowRaw(!showRaw)}
              className={`text-[10px] px-1.5 py-0.5 rounded transition-colors ${
                showRaw
                  ? "bg-muted text-primary"
                  : "text-muted-foreground hover:bg-muted"
              }`}
            >
              {showRaw ? "隐藏原始" : `+${stats.raw} 原始`}
            </button>
          )}
          <span className="ml-auto">
            <Button
              variant="ghost"
              size="sm"
              className="h-6 px-2 text-xs"
              onClick={handleReset}
            >
              <Undo2 className="mr-1 h-3 w-3" />
              重置
            </Button>
          </span>
        </div>

        {/* Search */}
        <div className="relative">
          <Search className="absolute left-2 top-1/2 h-3 w-3 -translate-y-1/2 text-muted-foreground" />
          <Input
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            placeholder="搜索名称或内容..."
            className="h-7 pl-7 text-xs"
          />
        </div>
      </div>

      <div className="grid gap-3 lg:grid-cols-[1fr_280px]">
        {/* Left: Element list */}
        <div className="h-[400px] overflow-auto rounded border">
          <div className="p-1">
            {filteredElements.length === 0 ? (
              <div className="flex h-20 items-center justify-center text-xs text-muted-foreground">
                {searchQuery ? "无匹配结果" : "该类型暂无元素"}
              </div>
            ) : (
              filteredElements.map((el) => {
                const isSelected = el.id === selectedId;
                return (
                  <button
                    key={el.id}
                    onClick={() =>
                      setSelectedId(isSelected ? null : el.id)
                    }
                    className={`flex w-full items-center gap-2 rounded px-2 py-1 text-left text-xs transition-colors ${
                      isSelected
                        ? "bg-primary/10 text-primary"
                        : "hover:bg-muted"
                    }`}
                  >
                    <ElementIcon type={el.type} />
                    <span className="flex-1 truncate font-mono">
                      {el.type === "node" && el.name && (
                        <span className="mr-1 text-muted-foreground">
                          ({el.name})
                        </span>
                      )}
                      {getLabel(el, nodeContentMap)}
                    </span>
                    {typeFilter === "all" && (
                      <span className="shrink-0 text-[10px] text-muted-foreground">
                        {TYPE_LABELS[el.type]}
                      </span>
                    )}
                  </button>
                );
              })
            )}
          </div>
        </div>

        {/* Right: Property panel */}
        <div className="rounded border p-3">
          {selectedElement ? (
            <>
              {selectedElement.type === "node" && (
                <NodePropertyPanel
                  node={selectedElement}
                  onChange={handleElementChange}
                />
              )}
              {selectedElement.type === "draw" && (
                <DrawPropertyPanel
                  draw={selectedElement}
                  onChange={handleElementChange}
                />
              )}
              {selectedElement.type === "raw" && (
                <RawPropertyPanel
                  el={selectedElement}
                  onChange={handleElementChange}
                />
              )}
              {selectedElement.type === "coordinate" && (
                <div className="space-y-3">
                  <h4 className="text-xs font-semibold text-muted-foreground tracking-wide">
                    坐标属性
                  </h4>
                  <div>
                    <label className="text-xs text-muted-foreground">
                      名称
                    </label>
                    <Input
                      value={selectedElement.name}
                      disabled
                      className="h-7 text-xs bg-muted"
                    />
                  </div>
                  <div>
                    <label className="text-xs text-muted-foreground">
                      位置
                    </label>
                    <Input
                      value={selectedElement.position}
                      onChange={(e) =>
                        handleElementChange({
                          ...selectedElement,
                          position: e.target.value,
                        })
                      }
                      className="h-7 text-xs"
                    />
                  </div>
                </div>
              )}
            </>
          ) : (
            <div className="flex h-full items-center justify-center text-xs text-muted-foreground">
              点击左侧元素进行编辑
            </div>
          )}
        </div>
      </div>

      {/* Render button */}
      <div className="flex items-center gap-2">
        <Button
          onClick={handleRender}
          disabled={isRendering}
          className="flex-1"
        >
          {isRendering ? (
            <Loader2 className="mr-2 h-4 w-4 animate-spin" />
          ) : (
            <RefreshCw className="mr-2 h-4 w-4" />
          )}
          {isRendering ? "渲染中..." : "应用修改并渲染"}
        </Button>
      </div>

      {renderError && (
        <div className="rounded border border-red-200 bg-red-50 p-2 text-xs text-red-700 dark:border-red-800 dark:bg-red-950 dark:text-red-300">
          {renderError}
        </div>
      )}
    </div>
  );
}
