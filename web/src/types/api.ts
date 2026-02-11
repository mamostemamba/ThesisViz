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
  explanation?: string;
  review_issues?: Record<string, unknown>;
  created_at: string;
}

export interface HealthResponse {
  status: string;
  details: Record<string, string>;
}

export interface RenderRequest {
  code: string;
  timeout?: number;
}

export interface RenderResponse {
  status: string;
  image?: string;
  error?: string;
}
