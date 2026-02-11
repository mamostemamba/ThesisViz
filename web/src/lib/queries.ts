import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import {
  listProjects,
  createProject,
  getProject,
  deleteProject,
  listGenerations,
  renderCode,
} from "./api";
import type {
  CreateProjectRequest,
  RenderRequest,
} from "@/types/api";

// Projects
export function useProjects(page = 1, pageSize = 20) {
  return useQuery({
    queryKey: ["projects", page, pageSize],
    queryFn: () => listProjects(page, pageSize),
  });
}

export function useProject(id: string) {
  return useQuery({
    queryKey: ["project", id],
    queryFn: () => getProject(id),
    enabled: !!id,
  });
}

export function useCreateProject() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateProjectRequest) => createProject(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["projects"] });
    },
  });
}

export function useDeleteProject() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => deleteProject(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["projects"] });
    },
  });
}

// Generations
export function useGenerations(projectId: string, page = 1, pageSize = 20) {
  return useQuery({
    queryKey: ["generations", projectId, page, pageSize],
    queryFn: () => listGenerations(projectId, page, pageSize),
    enabled: !!projectId,
  });
}

// Render
export function useRender() {
  return useMutation({
    mutationFn: (data: RenderRequest) => renderCode(data),
  });
}
