"""Python rendering sidecar service."""

from fastapi import FastAPI
from pydantic import BaseModel

from executor import render_matplotlib

app = FastAPI(title="ThesisViz Render Service")


@app.get("/health")
def health():
    return {"status": "ok"}


class RenderRequest(BaseModel):
    code: str
    timeout: int = 30


@app.post("/render/matplotlib")
def matplotlib_render(req: RenderRequest):
    result = render_matplotlib(req.code, timeout=req.timeout)
    if "error" in result:
        return {"status": "error", "error": result["error"]}
    return {"status": "ok", "image": result["image"]}
