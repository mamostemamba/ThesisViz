"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { useProjects, useCreateProject, useDeleteProject } from "@/lib/queries";
import { ProjectCard } from "./ProjectCard";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Plus, Loader2 } from "lucide-react";

export function ProjectGallery() {
  const router = useRouter();
  const { data, isLoading } = useProjects();
  const createMutation = useCreateProject();
  const deleteMutation = useDeleteProject();
  const [newTitle, setNewTitle] = useState("");
  const [showInput, setShowInput] = useState(false);

  const handleCreate = async () => {
    const title = newTitle.trim() || "未命名项目";
    const project = await createMutation.mutateAsync({ title });
    setNewTitle("");
    setShowInput(false);
    router.push(`/project?id=${project.id}`);
  };

  const handleDelete = async (id: string) => {
    await deleteMutation.mutateAsync(id);
  };

  const projects = data?.items ?? [];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold tracking-tight">项目列表</h2>
          <p className="text-sm text-muted-foreground">
            创建和管理你的学术图表生成项目
          </p>
        </div>
        {!showInput && (
          <Button onClick={() => setShowInput(true)}>
            <Plus className="mr-2 h-4 w-4" />
            新建项目
          </Button>
        )}
      </div>

      {showInput && (
        <div className="flex gap-2">
          <Input
            placeholder="输入项目标题..."
            value={newTitle}
            onChange={(e) => setNewTitle(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === "Enter") handleCreate();
              if (e.key === "Escape") setShowInput(false);
            }}
            autoFocus
          />
          <Button
            onClick={handleCreate}
            disabled={createMutation.isPending}
          >
            {createMutation.isPending ? (
              <Loader2 className="mr-2 h-4 w-4 animate-spin" />
            ) : null}
            创建
          </Button>
          <Button variant="outline" onClick={() => setShowInput(false)}>
            取消
          </Button>
        </div>
      )}

      {isLoading ? (
        <div className="flex items-center justify-center py-12">
          <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
        </div>
      ) : projects.length === 0 ? (
        <div className="flex flex-col items-center justify-center rounded-lg border border-dashed py-16">
          <p className="text-sm text-muted-foreground">
            暂无项目，创建一个开始使用吧
          </p>
          <Button
            className="mt-4"
            variant="outline"
            onClick={() => setShowInput(true)}
          >
            <Plus className="mr-2 h-4 w-4" />
            创建第一个项目
          </Button>
        </div>
      ) : (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {projects.map((project) => (
            <ProjectCard
              key={project.id}
              project={project}
              onDelete={handleDelete}
            />
          ))}
        </div>
      )}
    </div>
  );
}
