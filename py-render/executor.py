"""Matplotlib sandbox executor using multiprocessing isolation."""

import io
import re
import base64
import multiprocessing
from typing import Optional


def _strip_imports(code: str) -> str:
    """Remove import lines - the namespace already provides plt, np, matplotlib."""
    return re.sub(
        r"^(import\s+.+|from\s+\S+\s+import\s+.+)$", "", code, flags=re.MULTILINE
    )


def _execute_in_process(code: str, result_queue: multiprocessing.Queue) -> None:
    """Run matplotlib code in an isolated process and put the result PNG bytes on the queue."""
    import matplotlib

    matplotlib.use("Agg")
    import matplotlib.pyplot as plt
    import numpy as np

    try:
        plt.close("all")
        clean_code = _strip_imports(code)

        namespace = {
            "plt": plt,
            "np": np,
            "numpy": np,
            "matplotlib": matplotlib,
            "__builtins__": {
                "range": range,
                "len": len,
                "enumerate": enumerate,
                "zip": zip,
                "list": list,
                "dict": dict,
                "tuple": tuple,
                "set": set,
                "int": int,
                "float": float,
                "str": str,
                "bool": bool,
                "min": min,
                "max": max,
                "sum": sum,
                "abs": abs,
                "round": round,
                "sorted": sorted,
                "reversed": reversed,
                "print": print,
                "map": map,
                "filter": filter,
                "True": True,
                "False": False,
                "None": None,
            },
        }

        exec(clean_code, namespace)  # noqa: S102
        fig = plt.gcf()
        if not fig.get_axes():
            result_queue.put({"error": "No axes found in figure"})
            return

        buf = io.BytesIO()
        fig.savefig(buf, format="png", dpi=150, bbox_inches="tight")
        buf.seek(0)
        result_queue.put({"image": base64.b64encode(buf.read()).decode()})
    except Exception as e:
        result_queue.put({"error": str(e)})
    finally:
        plt.close("all")


def render_matplotlib(code: str, timeout: int = 30) -> dict:
    """Execute matplotlib code in an isolated process.

    Returns dict with either 'image' (base64 PNG) or 'error' key.
    """
    queue: multiprocessing.Queue = multiprocessing.Queue()
    proc = multiprocessing.Process(target=_execute_in_process, args=(code, queue))
    proc.start()
    proc.join(timeout=timeout)

    if proc.is_alive():
        proc.kill()
        proc.join()
        return {"error": f"Execution timed out after {timeout}s"}

    if queue.empty():
        return {"error": "Process terminated without producing output"}

    return queue.get()
