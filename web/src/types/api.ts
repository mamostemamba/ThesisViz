export interface Project {
  id: string;
  title: string;
  settings?: Record<string, unknown>;
  created_at: string;
}

export interface Generation {
  id: string;
  project_id: string;
  parent_id?: string;
  format: "tikz" | "matplotlib" | "mermaid";
  prompt: string;
  status: "queued" | "processing" | "success" | "failed";
  code?: string;
  image_key?: string;
  image_url?: string;
  explanation?: string;
  review_issues?: Record<string, unknown>;
  created_at: string;
}

export interface HealthResponse {
  status: string;
  details: Record<string, string>;
}

export interface PaginatedResponse<T> {
  items: T[];
  total: number;
  page: number;
  page_size: number;
}

export interface RenderRequest {
  code: string;
  format: string;
  language?: string;
  color_scheme?: string;
  generation_id?: string;
  dpi?: number;
  timeout?: number;
}

export interface RenderResponse {
  status: string;
  image_url?: string;
  image_key?: string;
  error?: string;
}

export interface CreateProjectRequest {
  title: string;
  settings?: string;
}

export interface UpdateProjectRequest {
  title?: string;
  settings?: string;
}

export interface CreateGenerationRequest {
  format: string;
  prompt: string;
  parent_id?: string;
  code?: string;
}
