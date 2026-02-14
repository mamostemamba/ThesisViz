<div align="center">

# <img src="https://img.icons8.com/fluency/48/graph-report.png" width="32" height="32" alt="logo" /> ThesisViz

**AI-Powered Academic Diagram Generator**

Describe in natural language, get publication-ready vector graphics

[![Go](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat-square&logo=go&logoColor=white)](https://go.dev)
[![Next.js](https://img.shields.io/badge/Next.js-16-000000?style=flat-square&logo=next.js&logoColor=white)](https://nextjs.org)
[![React](https://img.shields.io/badge/React-19-61DAFB?style=flat-square&logo=react&logoColor=black)](https://react.dev)
[![Python](https://img.shields.io/badge/Python-3.11+-3776AB?style=flat-square&logo=python&logoColor=white)](https://python.org)
[![Gemini](https://img.shields.io/badge/Gemini_API-4285F4?style=flat-square&logo=google&logoColor=white)](https://ai.google.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow?style=flat-square)](LICENSE)

[中文](./README.md) | English

</div>

<div align="center">

<table>
<tr>
<td align="center"><b>📝 Paste a paper paragraph — AI recommends diagrams</b><br><br><img src="docs/screenshots/1-analyze.png" width="480" alt="AI Analysis" /></td>
<td align="center"><b>⚙️ Auto-generated instructions + real-time progress</b><br><br><img src="docs/screenshots/2-generate.png" width="480" alt="Generation Progress" /></td>
</tr>
<tr>
<td align="center"><b>🖼️ Publication-ready diagram + source code</b><br><br><img src="docs/screenshots/3-result.png" width="480" alt="Result" /></td>
<td align="center"><b>🔄 Auto-redraw + AI visual review iterations</b><br><br><img src="docs/screenshots/4-review.png" width="480" alt="Review & Redraw" /></td>
</tr>
<tr>
<td align="center" colspan="2"><b>✏️ TikZ Fine-Tuning Editor — visual element selection with precise property editing</b><br><br><img src="docs/screenshots/5-finetune.png" width="700" alt="Fine-Tune Editor" /></td>
</tr>
</table>

</div>

---

## Table of Contents

- [ ThesisViz](#-thesisviz)
  - [Table of Contents](#table-of-contents)
  - [✨ Features](#-features)
    - [🎯 Natural Language → Diagrams](#-natural-language--diagrams)
    - [🖼️ Multiple Output Formats](#️-multiple-output-formats)
    - [🔄 Conversational Refinement](#-conversational-refinement)
    - [👁️ AI Visual Review](#️-ai-visual-review)
    - [🎨 Smart Color Schemes](#-smart-color-schemes)
    - [📤 One-Click Export](#-one-click-export)
    - [✏️ TikZ Fine-Tuning Editor](#️-tikz-fine-tuning-editor)
  - [🏗️ Architecture](#️-architecture)
  - [🚀 Quick Start](#-quick-start)
    - [Prerequisites](#prerequisites)
    - [Step by Step](#step-by-step)
  - [📖 Workflow](#-workflow)
  - [🔑 API Key Configuration](#-api-key-configuration)
  - [📋 Environment Variables](#-environment-variables)
  - [📁 Project Structure](#-project-structure)
  - [🐳 Docker Deployment](#-docker-deployment)
  - [❓ FAQ](#-faq)
  - [🛠️ Tech Stack](#️-tech-stack)
  - [🙏 Acknowledgements](#-acknowledgements)
  - [📄 License](#-license)

## ✨ Features

<table>
<tr>
<td width="50%">

### 🎯 Natural Language → Diagrams
Paste a paragraph from your paper or describe what you need. The AI analyzes your text and recommends suitable diagram types — no coding required.

### 🖼️ Multiple Output Formats
- **TikZ** — Publication-quality vector graphics for LaTeX papers
- **Matplotlib** — Data visualization charts
- **Mermaid** — Flowcharts, sequence diagrams, swimlane diagrams

### 🔄 Conversational Refinement
Describe modifications in natural language. The AI iterates on the existing code to apply your changes.

</td>
<td width="50%">

### 👁️ AI Visual Review
Gemini Vision automatically inspects generated diagrams for quality issues and fixes them — no manual intervention needed.

### 🎨 Smart Color Schemes
Built-in academic color palettes, plus the ability to extract custom colors from any image.

### 📤 One-Click Export
Export complete `.tex` files ready for Overleaf, or download PNG images.

### ✏️ TikZ Fine-Tuning Editor
Visual element highlighting helps you locate and manually adjust code with precision.

</td>
</tr>
</table>

## 🏗️ Architecture

| Component | Responsibility |
|---|---|
| **Next.js Frontend** | UI, real-time progress display, code editor, diagram preview |
| **Go API Backend** | AI Agent orchestration, WebSocket streaming, render scheduling, persistence |
| **py-render Sidecar** | Matplotlib sandbox execution (process isolation + restricted builtins) |
| **PostgreSQL** | Projects and generation records |
| **Redis** | Caching (optional) |
| **MinIO** | S3-compatible object storage for generated images |

## 🚀 Quick Start

### Prerequisites

| Dependency | Version | Installation |
|---|---|---|
| Docker & Compose | - | [Download](https://docs.docker.com/get-docker/) |
| Go | 1.23+ | [Download](https://go.dev/dl/) |
| Node.js | 20+ | [Download](https://nodejs.org/) |
| Python | 3.11+ | [Download](https://www.python.org/downloads/) |
| TeX distribution | - | macOS: `brew install --cask mactex`<br>Ubuntu: `sudo apt install texlive-full`<br>Windows: [MiKTeX](https://miktex.org/download) |
| Gemini API Key | - | [Get for free →](https://aistudio.google.com/apikey) |

### Step by Step

**① Clone the repository**

```bash
git clone https://github.com/your-username/ThesisViz.git
cd ThesisViz
```

**② Configure environment variables**

```bash
cp .env.example .env
# Edit .env and add your Gemini API Key (or configure it later via the web UI)
```

**③ Start infrastructure** (PostgreSQL + Redis + MinIO)

```bash
make infra
```

**④ Install Python dependencies** (first time only)

```bash
make render-setup
```

**⑤ Start all services**

```bash
make dev
```

**⑥ Open your browser** → [http://localhost:3000](http://localhost:3000) 🎉

> 💡 You can also start services separately in three terminals: `make api`, `make render`, `make web`

## 📖 Workflow

```
  📝 Paste paper paragraph
       │
       ▼
  🤖 AI analysis → recommends 3 diagram options
       │
       ▼
  🎯 Select an option → AI generates drawing instructions
       │
       ▼
  ⚙️ Automated pipeline: generate code → compile → visual review → auto-fix
       │
       ▼
  🖼️ Get diagram → conversational refinement → export when satisfied
       │
       ▼
  📤 Download PNG / Export .tex for Overleaf
```

1. **Create a project** — Create a project for your paper with title and abstract for context
2. **Input text** — Paste the paragraph that needs a diagram
3. **AI analysis** — The system analyzes your text and recommends 3 diagram options
4. **Select an option** — Pick one, and the AI generates detailed drawing instructions
5. **Generate diagram** — The automated pipeline streams real-time progress to your browser
6. **Iterate** — Describe modifications in natural language; the AI refines the existing code
7. **Export** — Download PNG or export `.tex` for Overleaf

## 🔑 API Key Configuration

ThesisViz requires a Google Gemini API Key. Two ways to configure:

<table>
<tr>
<td width="50%">

**🌐 Web UI (recommended for local use)**

Open the home page after starting. Enter your API Key in the card at the top and click Save.

- Key stored in browser localStorage only
- Automatically restored on page refresh
- Never sent to any third-party server

</td>
<td width="50%">

**⚙️ Environment variable (recommended for deployment)**

Set in your `.env` file:

```env
GEMINI_API_KEY=AIzaSy...
```

Loaded automatically at startup — no web configuration needed.

</td>
</tr>
</table>

> 🔗 Get a free API Key from [Google AI Studio](https://aistudio.google.com/apikey)

## 📋 Environment Variables

| Variable | Default | Description |
|---|---|---|
| `GEMINI_API_KEY` | — | Gemini API key (can also be set via web UI) |
| `GEMINI_MODEL` | `gemini-3-pro-preview` | Default LLM model |
| `DB_URL` | `postgres://thesisviz:...` | PostgreSQL connection string |
| `REDIS_URL` | `redis://localhost:6379/0` | Redis connection string |
| `MINIO_ENDPOINT` | `localhost:9000` | MinIO endpoint |
| `MINIO_ACCESS_KEY` | `minioadmin` | MinIO access key |
| `MINIO_SECRET_KEY` | `minioadmin` | MinIO secret key |
| `MINIO_BUCKET` | `thesisviz` | MinIO bucket name |
| `GO_API_PORT` | `8080` | Go API port |
| `PY_RENDER_URL` | `http://localhost:8081` | Python render service URL |

> See [`.env.example`](.env.example) for all defaults.

## 📁 Project Structure

```
ThesisViz/
├── go-api/                    # 🔧 Go Backend
│   ├── cmd/server/            #    Entry point
│   ├── internal/
│   │   ├── agent/             #    AI Agents (Router / TikZ / Matplotlib / Mermaid)
│   │   ├── handler/           #    HTTP & WebSocket handlers
│   │   ├── llm/               #    Gemini SDK wrapper
│   │   ├── prompt/            #    Prompt templates
│   │   ├── renderer/          #    TikZ compiler (pdflatex / xelatex)
│   │   ├── service/           #    Business logic & pipeline orchestration
│   │   └── ws/                #    WebSocket Hub
│   └── pkg/                   #    Shared packages (color schemes / code sanitizer)
│
├── py-render/                 # 🐍 Python Render Sidecar
│   ├── main.py                #    FastAPI entry point
│   └── executor.py            #    Matplotlib sandbox executor
│
├── web/                       # 🌐 Next.js Frontend
│   └── src/
│       ├── app/               #    Page routes
│       ├── components/        #    UI components (Shadcn/ui)
│       ├── lib/               #    API client & utilities
│       └── stores/            #    Zustand state management
│
├── deploy/docker/             # 🐳 Dockerfiles (go-api / py-render)
├── docker-compose.yml         #    Infrastructure (PostgreSQL / Redis / MinIO)
├── Makefile                   #    Dev commands
└── .env.example               #    Environment variable template
```

## 🐳 Docker Deployment

Production-ready multi-stage Dockerfiles are provided:

```bash
# Build Go API image
docker build -f deploy/docker/go-api.Dockerfile -t thesisviz-api .

# Build Python render service image
docker build -f deploy/docker/py-render.Dockerfile -t thesisviz-render .
```

Infrastructure is managed via Docker Compose:

```bash
# Start PostgreSQL + Redis + MinIO
docker compose up -d

# Stop
docker compose down
```

## ❓ FAQ

<details>
<summary><b>TikZ diagrams fail to compile?</b></summary>

Make sure you have a full TeX distribution installed:
- macOS → MacTeX (`brew install --cask mactex`)
- Linux → texlive-full (`sudo apt install texlive-full`)
- Windows → [MiKTeX](https://miktex.org/download)

For Chinese labels, `xelatex` is required (included in MacTeX and texlive-full).
</details>

<details>
<summary><b>Chinese text in Matplotlib shows as boxes?</b></summary>

The render service automatically uses Arial Unicode MS. If your system doesn't have this font, install any CJK-capable font (e.g., Noto Sans CJK).
</details>

<details>
<summary><b>Is my API Key secure?</b></summary>

API Keys configured via the web UI are stored only in your browser's localStorage and sent to **your own** backend service. The backend calls the Google API directly — no third parties involved.
</details>

<details>
<summary><b>Can I skip Docker?</b></summary>

Docker is only used for infrastructure (PostgreSQL, Redis, MinIO). You can install these services manually and update the connection strings in `.env`.
</details>

<details>
<summary><b>Which Gemini models are supported?</b></summary>

The default is `gemini-3-pro-preview`. You can switch models in the sidebar, or change the default via the `GEMINI_MODEL` environment variable.
</details>

## 🛠️ Tech Stack

<table>
<tr>
<td align="center" width="110">
<img src="https://cdn.jsdelivr.net/gh/devicons/devicon/icons/go/go-original-wordmark.svg" width="40" height="40" alt="Go" /><br><sub>Go</sub>
</td>
<td align="center" width="110">
<img src="https://cdn.jsdelivr.net/gh/devicons/devicon/icons/nextjs/nextjs-original.svg" width="40" height="40" alt="Next.js" /><br><sub>Next.js</sub>
</td>
<td align="center" width="110">
<img src="https://cdn.jsdelivr.net/gh/devicons/devicon/icons/react/react-original.svg" width="40" height="40" alt="React" /><br><sub>React</sub>
</td>
<td align="center" width="110">
<img src="https://cdn.jsdelivr.net/gh/devicons/devicon/icons/python/python-original.svg" width="40" height="40" alt="Python" /><br><sub>Python</sub>
</td>
<td align="center" width="110">
<img src="https://cdn.jsdelivr.net/gh/devicons/devicon/icons/postgresql/postgresql-original.svg" width="40" height="40" alt="PostgreSQL" /><br><sub>PostgreSQL</sub>
</td>
<td align="center" width="110">
<img src="https://cdn.jsdelivr.net/gh/devicons/devicon/icons/redis/redis-original.svg" width="40" height="40" alt="Redis" /><br><sub>Redis</sub>
</td>
<td align="center" width="110">
<img src="https://cdn.jsdelivr.net/gh/devicons/devicon/icons/docker/docker-original.svg" width="40" height="40" alt="Docker" /><br><sub>Docker</sub>
</td>
</tr>
</table>

## 🙏 Acknowledgements

- [Google Gemini API](https://ai.google.dev/) — AI generation & visual review
- [Gin](https://github.com/gin-gonic/gin) — Go web framework
- [Next.js](https://nextjs.org/) — React full-stack framework
- [Shadcn/ui](https://ui.shadcn.com/) — UI component library
- [TikZ / PGF](https://tikz.dev/) — TeX vector graphics system
- [Matplotlib](https://matplotlib.org/) — Python data visualization

## 📄 License

[MIT](LICENSE) — free to use, modify, and distribute.

---

<div align="center">

**If you find this useful, please give it a ⭐ Star!**

Made with ❤️ for researchers

</div>
