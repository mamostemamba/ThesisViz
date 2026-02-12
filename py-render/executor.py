"""Matplotlib sandbox executor using multiprocessing isolation."""

import io
import re
import base64
import multiprocessing
from typing import Optional

# Verified CJK fonts — only include fonts confirmed to exist.
# Arial Unicode MS has full CJK coverage and is available on macOS.
# DejaVu Sans is the matplotlib default fallback for non-CJK.
_CJK_FONTS = ["Arial Unicode MS", "DejaVu Sans"]


def _strip_imports(code: str) -> str:
    """Remove import lines and font overrides — managed by the executor."""
    # Strip imports
    code = re.sub(
        r"^(import\s+.+|from\s+\S+\s+import\s+.+)$", "", code, flags=re.MULTILINE
    )
    # Strip ALL rcParams font/unicode_minus overrides
    code = re.sub(
        r"^.*rcParams\s*\[\s*['\"]font\..+$", "", code, flags=re.MULTILINE
    )
    code = re.sub(
        r"^.*rcParams\s*\[\s*['\"]axes\.unicode_minus.+$", "", code, flags=re.MULTILINE
    )
    # Strip inline font arguments: fontfamily='...', fontname='...'
    code = re.sub(
        r",?\s*(?:fontfamily|fontname|font_family)\s*=\s*['\"][^'\"]*['\"]", "", code
    )
    return code


def _apply_cjk_rcparams(matplotlib_mod):
    """Set rcParams to use CJK fonts."""
    matplotlib_mod.rcParams["font.sans-serif"] = list(_CJK_FONTS)
    matplotlib_mod.rcParams["font.family"] = "sans-serif"
    matplotlib_mod.rcParams["axes.unicode_minus"] = False


def _force_cjk_fonts(fig):
    """Force CJK-capable fonts on EVERY text object in the figure."""
    import matplotlib.text as mtext
    for text_obj in fig.findobj(mtext.Text):
        text_obj.set_fontfamily(_CJK_FONTS)


def _execute_in_process(code: str, result_queue: multiprocessing.Queue) -> None:
    """Run matplotlib code in an isolated process and put the result PNG bytes on the queue."""
    import matplotlib

    matplotlib.use("Agg")
    import matplotlib.pyplot as plt
    import numpy as np

    import matplotlib.patches as mpatches
    import matplotlib.colors as mcolors
    import matplotlib.ticker as mticker
    import matplotlib.gridspec as gridspec
    import matplotlib.patheffects as patheffects
    import matplotlib.collections as mcollections

    try:
        plt.close("all")

        # Pre-configure CJK fonts
        _apply_cjk_rcparams(matplotlib)

        # Monkey-patch plt.style.use so it doesn't reset fonts to Arial
        _original_style_use = plt.style.use
        def _safe_style_use(style):
            _original_style_use(style)
            _apply_cjk_rcparams(matplotlib)
        plt.style.use = _safe_style_use

        clean_code = _strip_imports(code)

        namespace = {
            "plt": plt,
            "np": np,
            "numpy": np,
            "matplotlib": matplotlib,
            "patches": mpatches,
            "mpatches": mpatches,
            "mcolors": mcolors,
            "mticker": mticker,
            "gridspec": gridspec,
            "patheffects": patheffects,
            "mcollections": mcollections,
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

        # Re-apply CJK rcParams after exec
        _apply_cjk_rcparams(matplotlib)

        fig = plt.gcf()
        if not fig.get_axes():
            result_queue.put({"error": "No axes found in figure"})
            return

        # FORCE CJK fonts on every Text object in the figure
        _force_cjk_fonts(fig)

        buf = io.BytesIO()
        fig.savefig(buf, format="png", dpi=150, bbox_inches="tight")
        buf.seek(0)
        result_queue.put({"image": base64.b64encode(buf.read()).decode()})
    except Exception as e:
        result_queue.put({"error": str(e)})
    finally:
        plt.close("all")


def render_matplotlib(code: str, timeout: int = 120) -> dict:
    """Execute matplotlib code in an isolated process.

    Returns dict with either 'image' (base64 PNG) or 'error' key.
    """
    queue: multiprocessing.Queue = multiprocessing.Queue()
    proc = multiprocessing.Process(target=_execute_in_process, args=(code, queue))
    proc.start()

    # Read from queue BEFORE joining — avoids deadlock when the child
    # puts a large payload (base64 image) that fills the pipe buffer.
    try:
        result = queue.get(timeout=timeout)
    except Exception:
        proc.kill()
        proc.join()
        return {"error": f"Execution timed out after {timeout}s"}

    proc.join(timeout=5)
    if proc.is_alive():
        proc.kill()
        proc.join()

    return result
