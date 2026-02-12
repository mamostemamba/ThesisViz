"use client";

import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Badge } from "@/components/ui/badge";
import { useSettingsStore } from "@/stores/useSettingsStore";
import { useGenerateStore } from "@/stores/useGenerateStore";

export function Sidebar() {
  const format = useSettingsStore((s) => s.format);
  const setFormat = useSettingsStore((s) => s.setFormat);
  const language = useSettingsStore((s) => s.language);
  const setLanguage = useSettingsStore((s) => s.setLanguage);
  const colorScheme = useSettingsStore((s) => s.colorScheme);
  const setColorScheme = useSettingsStore((s) => s.setColorScheme);
  const model = useSettingsStore((s) => s.model);
  const setModel = useSettingsStore((s) => s.setModel);
  const isRendering = useGenerateStore((s) => s.isRendering);

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
          <Select value={colorScheme} onValueChange={setColorScheme}>
            <SelectTrigger>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="drawio">Draw.io 经典</SelectItem>
              <SelectItem value="academic_blue">学术蓝</SelectItem>
              <SelectItem value="nature">Nature</SelectItem>
              <SelectItem value="ieee">IEEE</SelectItem>
            </SelectContent>
          </Select>
        </div>

        <div>
          <label className="mb-2 block text-sm font-medium">AI 模型</label>
          <Select value={model} onValueChange={setModel}>
            <SelectTrigger>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="gemini-2.5-flash">Gemini 2.5 Flash</SelectItem>
              <SelectItem value="gemini-2.5-pro">Gemini 2.5 Pro</SelectItem>
              <SelectItem value="gemini-2.0-flash">Gemini 2.0 Flash</SelectItem>
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
