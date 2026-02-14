/**
 * JSON-to-TikZ Re-Compiler
 *
 * Compiles a TikZDocument back into valid TikZ code.
 * For parsed elements (node, draw, coordinate), rebuilds from properties.
 * For raw elements, outputs content as-is for round-trip fidelity.
 */

import type {
  TikZDocument,
  TikZElement,
  TikZNode,
  TikZDraw,
  TikZCoordinate,
  TikZOption,
} from "./types";

/**
 * Compile a TikZOption[] array into a comma-separated string (without brackets).
 */
export function compileOptions(opts: TikZOption[]): string {
  return opts
    .map((o) => (o.value ? `${o.key}=${o.value}` : o.key))
    .join(", ");
}

/**
 * Compile a TikZNode back to a \node statement.
 */
function compileNode(node: TikZNode): string {
  const parts: string[] = ["\\node"];

  if (node.options.length > 0) {
    parts.push(`[${compileOptions(node.options)}]`);
  }

  if (node.name) {
    parts.push(` (${node.name})`);
  }

  if (node.position) {
    parts.push(` at (${node.position})`);
  }

  parts.push(` {${node.content}}`);
  parts.push(";");

  return parts.join("");
}

/**
 * Compile a TikZDraw back to a \draw statement.
 */
function compileDraw(draw: TikZDraw): string {
  const parts: string[] = [`\\${draw.command}`];

  if (draw.options.length > 0) {
    parts.push(`[${compileOptions(draw.options)}]`);
  }

  if (draw.path) {
    parts.push(` ${draw.path}`);
  }

  parts.push(";");
  return parts.join("");
}

/**
 * Compile a TikZCoordinate back to a \coordinate statement.
 */
function compileCoordinate(coord: TikZCoordinate): string {
  const parts: string[] = ["\\coordinate"];

  if (coord.name) {
    parts.push(` (${coord.name})`);
  }

  if (coord.position) {
    parts.push(` at (${coord.position})`);
  }

  parts.push(";");
  return parts.join("");
}

/**
 * Compile a single TikZElement back to TikZ code.
 */
function compileElement(el: TikZElement): string {
  switch (el.type) {
    case "node":
      return compileNode(el);
    case "draw":
      return compileDraw(el);
    case "coordinate":
      return compileCoordinate(el);
    case "raw":
      return el.content;
  }
}

/**
 * Insert a visual highlight overlay into the ORIGINAL TikZ code.
 *
 * Unlike compileTikZ() which reconstructs from the parsed model (lossy),
 * this function preserves the original code verbatim and only appends
 * a highlight overlay line before \end{tikzpicture}.
 *
 * For unnamed nodes it injects a temporary name into the raw statement.
 */
export function compileTikZWithHighlight(
  originalCode: string,
  doc: TikZDocument,
  elementId: string
): string {
  const target = doc.elements.find((e) => e.id === elementId);
  if (!target) return originalCode;

  let code = originalCode;
  let hlLine = "";

  if (target.type === "node") {
    let name = target.name;
    if (!name) {
      // Inject a temp name into the raw statement so we can reference the node
      const tempName = "__hl__";
      const raw = target.raw;
      if (raw) {
        let modified: string;
        const atIdx = raw.indexOf(" at ");
        if (atIdx !== -1) {
          // \node[opts] at (pos) {text} → \node[opts] (__hl__) at (pos) {text}
          modified = raw.slice(0, atIdx) + ` (${tempName})` + raw.slice(atIdx);
        } else {
          // \node[opts] {text} → \node[opts] (__hl__) {text}
          const braceIdx = raw.indexOf("{");
          if (braceIdx !== -1) {
            modified = raw.slice(0, braceIdx) + `(${tempName}) ` + raw.slice(braceIdx);
          } else {
            modified = raw;
          }
        }
        if (modified !== raw) {
          code = code.replace(raw, modified);
          name = tempName;
        }
      }
    }
    if (name) {
      hlLine = `    \\draw[red, line width=1.5pt, dashed, rounded corners=2pt] ([shift={(-2pt,-2pt)}]${name}.south west) rectangle ([shift={(2pt,2pt)}]${name}.north east);`;
    }
  } else if (target.type === "draw") {
    hlLine = `    \\${target.command}[red, line width=2.5pt, opacity=0.4] ${target.path};`;
  } else if (target.type === "coordinate" && target.name) {
    hlLine = `    \\fill[red, opacity=0.6] (${target.name}) circle (3pt);`;
  }

  if (hlLine) {
    code = code.replace(
      "\\end{tikzpicture}",
      `\n${hlLine}\n\\end{tikzpicture}`
    );
  }

  return code;
}

export function compileTikZ(doc: TikZDocument): string {
  const lines: string[] = [];

  // Opening
  if (doc.globalOptions.trim()) {
    lines.push(`\\begin{tikzpicture}[${doc.globalOptions}]`);
  } else {
    lines.push("\\begin{tikzpicture}");
  }

  // Elements
  lines.push("");
  for (const el of doc.elements) {
    const compiled = compileElement(el);
    if (compiled) {
      // Indent non-comment lines
      if (el.type === "raw" && el.content.startsWith("%")) {
        lines.push(`    ${compiled}`);
      } else if (el.type === "raw" && el.content === "") {
        lines.push("");
      } else {
        lines.push(`    ${compiled}`);
      }
    }
  }

  // Closing
  lines.push("");
  lines.push("\\end{tikzpicture}");

  return lines.join("\n");
}
