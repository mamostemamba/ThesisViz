# Project Blueprint: ThesisViz (v2.0)

## 1. 项目愿景

ThesisViz 是一个面向科研人员的**智能学术绘图平台**。它利用最新的 **Gemini 3 Pro** 多模态大模型，将自然语言指令或论文摘要转化为高质量的学术图表（TikZ 矢量图、Matplotlib 数据图、Mermaid 流程图）。

系统旨在解决科研绘图中“代码难写、调整繁琐、审美不统一”的痛点，提供**所见即所得**的实时预览、**基于视觉的自动审查**以及**多轮对话式修改**功能。

## 2. 系统架构 (Hybrid Architecture)

本项目采用**混合开发模式**，兼顾本地开发的灵活性与生产环境的稳定性。

### 2.1 拓扑结构

* **Host (本地宿主机)**: 运行核心业务代码（Next.js 前端、Go 后端、Python 渲染服务），方便调试。
* **Docker (容器设施)**: 运行数据库和中间件，确保环境纯净。

```mermaid
graph TD
    User[用户浏览器] -- HTTP/WebSocket --> NextJS[Next.js Frontend (Host)]
    
    subgraph "Host Machine (Mac/Linux)"
        NextJS -- API调用 --> GoAPI[Go Backend API]
        GoAPI -- 进程调用 --> Latex[本地 TeX Live]
        GoAPI -- HTTP调用 --> PySidecar[Python Renderer Service]
        
        GoAPI -- gRPC/REST --> Gemini[Google Gemini 3 Pro API]
    end
    
    subgraph "Docker Infrastructure"
        GoAPI -- TCP:5432 --> PG[(PostgreSQL)]
        GoAPI -- TCP:6379 --> Redis[(Redis MQ & Cache)]
        GoAPI -- TCP:9000 --> MinIO[(MinIO Object Storage)]
        
        Worker[Go Async Worker] -- 任务调度 --> Redis
    end

```

---

## 3. 核心功能模块

### 3.1 智能生成流水线 (The Pipeline)

不再是简单的请求-响应，而是一个**异步状态机**：

1. **Generate**: 调用 **Gemini 3 Pro** 生成初始代码。
2. **Render**:
* **TikZ**: 调用本地 `pdflatex` 编译 PDF -> 转 PNG。
* **Matplotlib**: 调用 Python Sidecar 沙箱执行 -> 获 PNG。
* **Mermaid**: 前端实时渲染。


3. **Auto-Fix (自我修复)**: 若编译/执行报错，自动将错误日志回传给 LLM 进行代码修正（最多重试 3 次）。
4. **Visual Review (视觉审查)**: 将生成的图片回传给 **Gemini 3 Pro Vision**，检查文字遮挡、布局溢出等问题。若未通过，自动触发修改。
5. **Explain**: 并行生成代码解释文档。

### 3.2 交互特性

* **实时反馈**: 全链路 WebSocket 推送，前端展示“生成中 -> 编译中 -> 审查中”的实时状态流。
* **对话式修改 (Refinement)**: 用户可以选中某张图，输入“把线条加粗”、“换成红色”，系统基于上下文进行增量修改。
* **在线编辑**: 自动生成 Overleaf (TikZ) 和 Mermaid Live 编辑链接，方便用户导出继续编辑。

---

## 4. 技术栈选型

### 4.1 Backend (Go)

* **Framework**: Gin (HTTP) + Gorilla WebSocket。
* **Task Queue**: Asynq (基于 Redis 的分布式任务队列)。
* **LLM SDK**: `google/generative-ai-go` (配置默认使用 `gemini-3-pro`，支持流式输出)。
* **Database**: PostgreSQL + GORM。
* **Storage**: MinIO (S3 兼容)。

### 4.2 Frontend (Next.js)

* **Framework**: Next.js 14 (App Router)。
* **UI**: Shadcn/ui + Tailwind CSS。
* **Editor**: Monaco Editor (VS Code 同款)，支持代码高亮与 Diff 对比。
* **State**: Zustand + React Query。

### 4.3 Rendering (Python Sidecar)

* **Framework**: FastAPI。
* **Security**: 使用 `multiprocessing` 独立进程执行 Matplotlib 代码，配合 `ast` 模块进行静态安全检查（禁止文件读写、网络请求）。

---

## 5. 数据库设计 (Schema)

```sql
-- 项目表
CREATE TABLE projects (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title           VARCHAR(255) NOT NULL,
    settings        JSONB, -- 存储语言(zh/en)、配色方案(academic_blue等)
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

-- 生成任务表 (支持链式修改)
CREATE TABLE generations (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID REFERENCES projects(id),
    parent_id       UUID REFERENCES generations(id), -- 指向前一个版本
    
    format          VARCHAR(20) NOT NULL, -- tikz, matplotlib, mermaid
    prompt          TEXT NOT NULL,
    status          VARCHAR(20) DEFAULT 'queued', -- queued, processing, success, failed
    
    code            TEXT,                 -- 最终代码
    image_key       VARCHAR(255),         -- MinIO 图片路径
    explanation     TEXT,                 -- 代码解释
    
    review_issues   JSONB,                -- 视觉审查出的问题
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

```

---

## 6. 开发环境配置清单

### 6.1 本地宿主机 (Host Requirements)

* **Go 1.23+**: 运行 Backend API & Worker。
* **Node.js 18+**: 运行 Frontend。
* **Python 3.10+**: 运行 Py-Render 服务。
* **TeX Live / BasicTeX**: 必须安装，用于 TikZ 编译。确保包含 `standalone`, `tikz`, `ctex` 包。

### 6.2 Docker Infrastructure

* `postgres:16-alpine`
* `redis:7-alpine`
* `minio/minio`

---

## 7. 实施计划 (Implementation Plan)

1. **Phase 1: 骨架搭建**
* 初始化 Next.js + Go Module + Python venv。
* 编写 `docker-compose.yml` 启动基础服务。


2. **Phase 2: 核心渲染引擎**
* 实现 Go 本地调用 `pdflatex`。
* 实现 Python Sidecar 安全执行 Matplotlib。


3. **Phase 3: 异步业务逻辑**
* 实现 Asynq 任务队列。
* 集成 Gemini 3 Pro SDK，实现“生成-编译-审查”状态机。


4. **Phase 4: 前端交互**
* 实现 WebSocket 实时日志组件。
* 集成 Monaco Editor 与图片预览分栏。
* 实现基于对话的修改功能。