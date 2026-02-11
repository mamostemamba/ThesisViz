"use client";

import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Badge } from "@/components/ui/badge";

export function Sidebar() {
  return (
    <aside className="w-64 border-r bg-muted/30 p-4">
      <div className="space-y-6">
        <div>
          <label className="mb-2 block text-sm font-medium">Format</label>
          <Select defaultValue="matplotlib">
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
          <label className="mb-2 block text-sm font-medium">Language</label>
          <Select defaultValue="en">
            <SelectTrigger>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="en">English</SelectItem>
              <SelectItem value="zh">Chinese</SelectItem>
            </SelectContent>
          </Select>
        </div>

        <div>
          <label className="mb-2 block text-sm font-medium">Color Scheme</label>
          <Select defaultValue="academic_blue">
            <SelectTrigger>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="academic_blue">Academic Blue</SelectItem>
              <SelectItem value="nature">Nature</SelectItem>
              <SelectItem value="ieee">IEEE</SelectItem>
              <SelectItem value="minimal">Minimal</SelectItem>
            </SelectContent>
          </Select>
        </div>

        <div>
          <label className="mb-2 block text-sm font-medium">Status</label>
          <Badge variant="secondary">Ready</Badge>
        </div>
      </div>
    </aside>
  );
}
