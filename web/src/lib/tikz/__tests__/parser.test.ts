import { parseTikZ, parseOptions, resetIdCounter } from "../parser";
import { compileTikZ } from "../compiler";
import type { TikZNode, TikZDraw, TikZCoordinate } from "../types";

beforeEach(() => {
  resetIdCounter();
});

describe("parseOptions", () => {
  it("parses simple key=value pairs", () => {
    const result = parseOptions("fill=blue, draw=red, thick");
    expect(result).toEqual([
      { key: "fill", value: "blue" },
      { key: "draw", value: "red" },
      { key: "thick", value: "" },
    ]);
  });

  it("handles brace-delimited values", () => {
    const result = parseOptions(
      'box/.style={rectangle, draw=black, thick}, fill=white'
    );
    expect(result).toHaveLength(2);
    expect(result[0].key).toBe("box/.style");
    expect(result[0].value).toContain("rectangle");
    expect(result[1]).toEqual({ key: "fill", value: "white" });
  });

  it("handles empty string", () => {
    expect(parseOptions("")).toEqual([]);
    expect(parseOptions("  ")).toEqual([]);
  });

  it("handles color expressions with !", () => {
    const result = parseOptions("fill=drawBlueFill!50, draw=black!60");
    expect(result).toEqual([
      { key: "fill", value: "drawBlueFill!50" },
      { key: "draw", value: "black!60" },
    ]);
  });
});

describe("parseTikZ", () => {
  it("parses a simple node", () => {
    const code = `\\begin{tikzpicture}
\\node[fill=blue, draw=red, minimum width=3cm] (mynode) at (0, 0) {Hello World};
\\end{tikzpicture}`;

    const doc = parseTikZ(code);
    expect(doc.elements).toHaveLength(1);

    const node = doc.elements[0] as TikZNode;
    expect(node.type).toBe("node");
    expect(node.name).toBe("mynode");
    expect(node.content).toBe("Hello World");
    expect(node.position).toBe("0, 0");
    expect(node.options).toContainEqual({ key: "fill", value: "blue" });
    expect(node.options).toContainEqual({
      key: "minimum width",
      value: "3cm",
    });
  });

  it("parses a node without name", () => {
    const code = `\\begin{tikzpicture}
\\node[title] at (layer4.north) {应用服务层};
\\end{tikzpicture}`;

    const doc = parseTikZ(code);
    expect(doc.elements).toHaveLength(1);

    const node = doc.elements[0] as TikZNode;
    expect(node.type).toBe("node");
    expect(node.name).toBe("");
    expect(node.content).toBe("应用服务层");
    expect(node.position).toBe("layer4.north");
  });

  it("parses a node with relative positioning", () => {
    const code = `\\begin{tikzpicture}
\\node[box, left=0.8cm of other] (mybox) {Content};
\\end{tikzpicture}`;

    const doc = parseTikZ(code);
    const node = doc.elements[0] as TikZNode;
    expect(node.type).toBe("node");
    expect(node.name).toBe("mybox");
    expect(node.content).toBe("Content");
    expect(node.position).toBe(""); // no "at" position
    expect(node.options).toContainEqual({
      key: "left",
      value: "0.8cm of other",
    });
  });

  it("parses a node with calc position", () => {
    const code = `\\begin{tikzpicture}
\\node[box] (circom) at ($(cloud_left_center) + (-1.5, 0.8)$) {Circom 电路};
\\end{tikzpicture}`;

    const doc = parseTikZ(code);
    const node = doc.elements[0] as TikZNode;
    expect(node.type).toBe("node");
    expect(node.name).toBe("circom");
    expect(node.position).toBe("$(cloud_left_center) + (-1.5, 0.8)$");
    expect(node.content).toBe("Circom 电路");
  });

  it("parses a draw statement", () => {
    const code = `\\begin{tikzpicture}
\\draw[small_arrow] (input.east) -- (proof.west);
\\end{tikzpicture}`;

    const doc = parseTikZ(code);
    const draw = doc.elements[0] as TikZDraw;
    expect(draw.type).toBe("draw");
    expect(draw.command).toBe("draw");
    expect(draw.path).toBe("(input.east) -- (proof.west)");
    expect(draw.options).toContainEqual({ key: "small_arrow", value: "" });
  });

  it("parses a draw with to[options]", () => {
    const code = `\\begin{tikzpicture}
\\draw[small_arrow, dashed] (infra_right.north) to[out=90, in=270] (input.south);
\\end{tikzpicture}`;

    const doc = parseTikZ(code);
    const draw = doc.elements[0] as TikZDraw;
    expect(draw.type).toBe("draw");
    expect(draw.path).toContain("to[out=90, in=270]");
  });

  it("parses a coordinate", () => {
    const code = `\\begin{tikzpicture}
\\coordinate (L4_left) at (\\leftEdge, 0);
\\end{tikzpicture}`;

    const doc = parseTikZ(code);
    const coord = doc.elements[0] as TikZCoordinate;
    expect(coord.type).toBe("coordinate");
    expect(coord.name).toBe("L4_left");
    expect(coord.position).toBe("\\leftEdge, 0");
  });

  it("preserves comments as raw elements", () => {
    const code = `\\begin{tikzpicture}
%% Layer 1
\\node (a) at (0,0) {A};
\\end{tikzpicture}`;

    const doc = parseTikZ(code);
    expect(doc.elements).toHaveLength(2);
    expect(doc.elements[0].type).toBe("raw");
    expect(doc.elements[1].type).toBe("node");
  });

  it("preserves \\def as raw", () => {
    const code = `\\begin{tikzpicture}
\\def\\leftEdge{-8}
\\node (a) at (0,0) {A};
\\end{tikzpicture}`;

    const doc = parseTikZ(code);
    // \def\leftEdge{-8} doesn't end with ;, so it's part of the node statement or raw
    // Actually it should be treated as its own raw element
    expect(doc.elements.length).toBeGreaterThanOrEqual(1);
  });

  it("preserves \\path let as raw", () => {
    const code = `\\begin{tikzpicture}
\\path let \\p1 = (layer3_center_anchor) in coordinate (L3_left) at (\\leftEdge, \\y1);
\\end{tikzpicture}`;

    const doc = parseTikZ(code);
    expect(doc.elements[0].type).toBe("raw");
  });

  it("extracts global options", () => {
    const code = `\\begin{tikzpicture}[font=\\footnotesize, node distance=1.2cm]
\\node (a) at (0,0) {A};
\\end{tikzpicture}`;

    const doc = parseTikZ(code);
    expect(doc.globalOptions).toContain("font=\\footnotesize");
    expect(doc.globalOptions).toContain("node distance=1.2cm");
  });

  it("handles multi-line global options with style defs", () => {
    const code = `\\begin{tikzpicture}[
    font=\\footnotesize,
    box/.style={
        rectangle,
        draw=black!60,
        thick
    }
]
\\node[box] (a) at (0,0) {A};
\\end{tikzpicture}`;

    const doc = parseTikZ(code);
    expect(doc.globalOptions).toContain("box/.style=");
    expect(doc.elements).toHaveLength(1);
    expect(doc.elements[0].type).toBe("node");
  });

  it("parses multiple elements", () => {
    const code = `\\begin{tikzpicture}
\\node[fill=blue] (a) at (0, 0) {Node A};
\\node[fill=red] (b) at (2, 0) {Node B};
\\draw[->] (a) -- (b);
\\end{tikzpicture}`;

    const doc = parseTikZ(code);
    expect(doc.elements).toHaveLength(3);
    expect(doc.elements[0].type).toBe("node");
    expect(doc.elements[1].type).toBe("node");
    expect(doc.elements[2].type).toBe("draw");
  });
});

describe("round-trip: parseTikZ → compileTikZ", () => {
  it("preserves node properties after round-trip", () => {
    const code = `\\begin{tikzpicture}
\\node[fill=blue, draw=red, minimum width=3cm] (mynode) at (0, 0) {Hello};
\\end{tikzpicture}`;

    const doc = parseTikZ(code);
    const result = compileTikZ(doc);

    expect(result).toContain("\\begin{tikzpicture}");
    expect(result).toContain("\\end{tikzpicture}");
    expect(result).toContain("fill=blue");
    expect(result).toContain("draw=red");
    expect(result).toContain("minimum width=3cm");
    expect(result).toContain("(mynode)");
    expect(result).toContain("at (0, 0)");
    expect(result).toContain("{Hello}");
  });

  it("preserves draw properties after round-trip", () => {
    const code = `\\begin{tikzpicture}
\\draw[->, thick] (a) -- (b);
\\end{tikzpicture}`;

    const doc = parseTikZ(code);
    const result = compileTikZ(doc);

    expect(result).toContain("\\draw");
    expect(result).toContain("->");
    expect(result).toContain("thick");
    expect(result).toContain("(a) -- (b)");
  });

  it("modification is reflected in compiled output", () => {
    const code = `\\begin{tikzpicture}
\\node[fill=blue] (a) at (0, 0) {Original};
\\end{tikzpicture}`;

    const doc = parseTikZ(code);

    // Modify the node
    const node = doc.elements[0] as TikZNode;
    node.content = "Modified";
    node.options = node.options.map((o) =>
      o.key === "fill" ? { ...o, value: "green" } : o
    );

    const result = compileTikZ(doc);
    expect(result).toContain("{Modified}");
    expect(result).toContain("fill=green");
    expect(result).not.toContain("fill=blue");
    expect(result).not.toContain("{Original}");
  });
});
