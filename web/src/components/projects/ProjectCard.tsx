"use client";

import { useRouter } from "next/navigation";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Trash2 } from "lucide-react";
import type { Project } from "@/types/api";

interface ProjectCardProps {
  project: Project;
  onDelete: (id: string) => void;
}

export function ProjectCard({ project, onDelete }: ProjectCardProps) {
  const router = useRouter();

  return (
    <Card
      className="cursor-pointer transition-colors hover:border-primary/50"
      onClick={() => router.push(`/project?id=${project.id}`)}
    >
      <CardHeader className="flex flex-row items-start justify-between space-y-0 pb-2">
        <CardTitle className="text-base font-medium line-clamp-1">
          {project.title}
        </CardTitle>
        <Button
          variant="ghost"
          size="icon"
          className="h-8 w-8 shrink-0 text-muted-foreground hover:text-destructive"
          onClick={(e) => {
            e.stopPropagation();
            onDelete(project.id);
          }}
        >
          <Trash2 className="h-4 w-4" />
        </Button>
      </CardHeader>
      <CardContent>
        <p className="text-xs text-muted-foreground">
          {new Date(project.created_at).toLocaleDateString()}
        </p>
      </CardContent>
    </Card>
  );
}
