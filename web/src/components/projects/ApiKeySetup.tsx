"use client";

import { useEffect, useState } from "react";
import { Card, CardContent } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { useSettingsStore } from "@/stores/useSettingsStore";
import { configureApiKey, getConfigStatus } from "@/lib/api";

export function ApiKeySetup() {
  const { apiKey, setApiKey } = useSettingsStore();
  const [inputKey, setInputKey] = useState(apiKey);
  const [showKey, setShowKey] = useState(false);
  const [hasKey, setHasKey] = useState<boolean | null>(null);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");
  const [collapsed, setCollapsed] = useState(false);

  // Check backend status and auto-send stored key on mount
  useEffect(() => {
    let cancelled = false;
    async function init() {
      try {
        const status = await getConfigStatus();
        if (cancelled) return;
        if (status.has_api_key) {
          setHasKey(true);
          setCollapsed(true);
          return;
        }
        // Backend has no key but we have one stored — send it
        if (apiKey) {
          await configureApiKey(apiKey);
          if (cancelled) return;
          setHasKey(true);
          setCollapsed(true);
        } else {
          setHasKey(false);
        }
      } catch {
        if (!cancelled) setHasKey(false);
      }
    }
    init();
    return () => { cancelled = true; };
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  async function handleSave() {
    if (!inputKey.trim()) return;
    setSaving(true);
    setError("");
    try {
      await configureApiKey(inputKey.trim());
      setApiKey(inputKey.trim());
      setHasKey(true);
      setCollapsed(true);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to set API key");
    } finally {
      setSaving(false);
    }
  }

  // Loading state
  if (hasKey === null) return null;

  // Collapsed — key is configured
  if (collapsed && hasKey) {
    return (
      <div className="mb-6 flex items-center gap-2 text-sm text-muted-foreground">
        <span className="inline-block h-2 w-2 rounded-full bg-green-500" />
        <span>API Key 已配置</span>
        <button
          onClick={() => setCollapsed(false)}
          className="ml-1 underline underline-offset-2 hover:text-foreground"
        >
          修改
        </button>
      </div>
    );
  }

  return (
    <Card className="mb-6">
      <CardContent className="flex flex-col gap-3">
        <div className="flex items-center gap-2">
          <span
            className={`inline-block h-2 w-2 rounded-full ${
              hasKey ? "bg-green-500" : "bg-yellow-500"
            }`}
          />
          <span className="text-sm font-medium">
            {hasKey ? "API Key 已配置" : "请配置 Gemini API Key"}
          </span>
        </div>
        <div className="flex gap-2">
          <div className="relative flex-1">
            <Input
              type={showKey ? "text" : "password"}
              placeholder="AIzaSy..."
              value={inputKey}
              onChange={(e) => setInputKey(e.target.value)}
              onKeyDown={(e) => e.key === "Enter" && handleSave()}
            />
            <button
              type="button"
              onClick={() => setShowKey(!showKey)}
              className="absolute right-2 top-1/2 -translate-y-1/2 text-xs text-muted-foreground hover:text-foreground"
            >
              {showKey ? "隐藏" : "显示"}
            </button>
          </div>
          <Button onClick={handleSave} disabled={saving || !inputKey.trim()}>
            {saving ? "保存中..." : "保存"}
          </Button>
        </div>
        {error && <p className="text-sm text-destructive">{error}</p>}
        <p className="text-xs text-muted-foreground">
          从{" "}
          <a
            href="https://aistudio.google.com/apikey"
            target="_blank"
            rel="noopener noreferrer"
            className="underline underline-offset-2 hover:text-foreground"
          >
            Google AI Studio
          </a>{" "}
          免费获取 API Key。密钥仅存储在浏览器本地，不会上传到任何第三方服务器。
        </p>
      </CardContent>
    </Card>
  );
}
