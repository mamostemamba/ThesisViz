"use client";

import { Textarea } from "@/components/ui/textarea";
import { useGenerateStore } from "@/stores/useGenerateStore";

export function CodeEditor() {
  const code = useGenerateStore((s) => s.code);
  const setCode = useGenerateStore((s) => s.setCode);

  return (
    <div className="flex h-full flex-col gap-2">
      <label className="text-sm font-medium">Code</label>
      <Textarea
        value={code}
        onChange={(e) => setCode(e.target.value)}
        placeholder={`Enter your TikZ or Matplotlib code here...\n\nExample TikZ:\n\\begin{tikzpicture}\n  \\node[draw,circle,fill=primaryFill,draw=primaryLine]{Hello};\n\\end{tikzpicture}\n\nExample Matplotlib:\nplt.figure()\nplt.plot([1,2,3],[1,4,9])\nplt.title("Test")`}
        className="min-h-0 flex-1 resize-none font-mono text-sm"
      />
    </div>
  );
}
