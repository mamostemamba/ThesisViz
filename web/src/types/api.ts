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
  status: "queued" | "processing" | "success" | "failed" | "cancelled";
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

export interface ColorPair {
  fill: string;
  line: string;
}

export interface CustomColors {
  pairs: ColorPair[];
}

export interface RenderRequest {
  code: string;
  format: string;
  language?: string;
  color_scheme?: string;
  custom_colors?: CustomColors;
  generation_id?: string;
  dpi?: number;
  timeout?: number;
  style?: string;
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

// AI Generation types

export interface AnalyzeRequest {
  text: string;
  language?: string;
  thesis_title?: string;
  thesis_abstract?: string;
  model?: string;
}

export interface Recommendation {
  title: string;
  description: string;
  drawing_prompt?: string;
  format?: string;
  priority: number;
  identity?: string;
}

export interface DrawingPromptRequest {
  text: string;
  title: string;
  description: string;
  identity?: string;
  language?: string;
  thesis_title?: string;
  thesis_abstract?: string;
  model?: string;
  color_scheme?: string;
  custom_colors?: CustomColors;
}

export interface DrawingPromptResponse {
  drawing_prompt: string;
}

export interface AnalyzeResponse {
  recommendations: Recommendation[];
}

export interface GenerateCreateRequest {
  project_id?: string;
  format: string;
  prompt: string;
  language?: string;
  color_scheme?: string;
  custom_colors?: CustomColors;
  thesis_title?: string;
  thesis_abstract?: string;
  model?: string;
  identity?: string;
  style?: string;
}

export interface GenerateCreateResponse {
  task_id: string;
}

export interface GenerateRefineRequest {
  generation_id: string;
  modification: string;
  language?: string;
  color_scheme?: string;
  custom_colors?: CustomColors;
  model?: string;
}

export interface GenerateRefineResponse {
  task_id: string;
}

export interface ExtractColorsResponse {
  colors: CustomColors;
}
