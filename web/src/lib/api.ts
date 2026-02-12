import type {
  Project,
  Generation,
  PaginatedResponse,
  RenderRequest,
  RenderResponse,
  CreateProjectRequest,
  UpdateProjectRequest,
  CreateGenerationRequest,
  AnalyzeRequest,
  AnalyzeResponse,
  GenerateCreateRequest,
  GenerateCreateResponse,
  GenerateRefineRequest,
  GenerateRefineResponse,
} from "@/types/api";

const BASE_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

export async function apiFetch<T>(
  path: string,
  options?: RequestInit
): Promise<T> {
  const res = await fetch(`${BASE_URL}${path}`, {
    headers: {
      "Content-Type": "application/json",
      ...options?.headers,
    },
    ...options,
  });

  if (!res.ok) {
    const body = await res.text();
    throw new Error(`API error ${res.status}: ${body}`);
  }

  return res.json();
}

// Health
export async function healthCheck() {
  return apiFetch<{ status: string; details: Record<string, string> }>(
    "/api/v1/health"
  );
}

// Projects
export async function createProject(data: CreateProjectRequest) {
  return apiFetch<Project>("/api/v1/projects", {
    method: "POST",
    body: JSON.stringify(data),
  });
}

export async function listProjects(page = 1, pageSize = 20) {
  return apiFetch<PaginatedResponse<Project>>(
    `/api/v1/projects?page=${page}&page_size=${pageSize}`
  );
}

export async function getProject(id: string) {
  return apiFetch<Project>(`/api/v1/projects/${id}`);
}

export async function updateProject(id: string, data: UpdateProjectRequest) {
  return apiFetch<Project>(`/api/v1/projects/${id}`, {
    method: "PUT",
    body: JSON.stringify(data),
  });
}

export async function deleteProject(id: string) {
  return apiFetch<{ message: string }>(`/api/v1/projects/${id}`, {
    method: "DELETE",
  });
}

// Generations
export async function createGeneration(
  projectId: string,
  data: CreateGenerationRequest
) {
  return apiFetch<Generation>(`/api/v1/projects/${projectId}/generations`, {
    method: "POST",
    body: JSON.stringify(data),
  });
}

export async function listGenerations(
  projectId: string,
  page = 1,
  pageSize = 20
) {
  return apiFetch<PaginatedResponse<Generation>>(
    `/api/v1/projects/${projectId}/generations?page=${page}&page_size=${pageSize}`
  );
}

export async function getGeneration(id: string) {
  return apiFetch<Generation>(`/api/v1/generations/${id}`);
}

export async function deleteGeneration(id: string) {
  return apiFetch<{ message: string }>(`/api/v1/generations/${id}`, {
    method: "DELETE",
  });
}

// Render
export async function renderCode(data: RenderRequest) {
  return apiFetch<RenderResponse>("/api/v1/render", {
    method: "POST",
    body: JSON.stringify(data),
  });
}

// Export full .tex for Overleaf
export async function exportTeX(data: {
  code: string;
  language?: string;
  color_scheme?: string;
}) {
  return apiFetch<{ tex: string }>("/api/v1/export/tex", {
    method: "POST",
    body: JSON.stringify(data),
  });
}

// AI Generate
export async function analyzeText(data: AnalyzeRequest) {
  return apiFetch<AnalyzeResponse>("/api/v1/generate/analyze", {
    method: "POST",
    body: JSON.stringify(data),
  });
}

export async function generateCreate(data: GenerateCreateRequest) {
  return apiFetch<GenerateCreateResponse>("/api/v1/generate/create", {
    method: "POST",
    body: JSON.stringify(data),
  });
}

export async function generateRefine(data: GenerateRefineRequest) {
  return apiFetch<GenerateRefineResponse>("/api/v1/generate/refine", {
    method: "POST",
    body: JSON.stringify(data),
  });
}

export async function getGenerationDetail(id: string) {
  return apiFetch<Generation>(`/api/v1/generate/${id}`);
}
