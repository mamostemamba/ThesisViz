/** TikZ object model for the post-processing editor. */

export interface TikZDocument {
  /** Raw global options from \begin{tikzpicture}[...] (without brackets) */
  globalOptions: string;
  /** Parsed elements inside the tikzpicture body */
  elements: TikZElement[];
}

export type TikZElement = TikZNode | TikZDraw | TikZCoordinate | TikZRaw;

export interface TikZNode {
  type: "node";
  id: string;
  /** TikZ node name, e.g. "app_mini" from (app_mini) */
  name: string;
  /** Parsed key-value options */
  options: TikZOption[];
  /** Raw position string, e.g. "0, 0" or "$(node) + (1,2)$" */
  position: string;
  /** Text content inside {} */
  content: string;
  /** Original raw statement for fallback */
  raw: string;
}

export interface TikZDraw {
  type: "draw";
  id: string;
  /** Command: draw, fill, filldraw, path */
  command: string;
  /** Parsed key-value options */
  options: TikZOption[];
  /** Full path specification after options */
  path: string;
  raw: string;
}

export interface TikZCoordinate {
  type: "coordinate";
  id: string;
  name: string;
  /** Raw position string */
  position: string;
  raw: string;
}

export interface TikZRaw {
  type: "raw";
  id: string;
  /** Unparsed content (comments, \def, \path, scope blocks, etc.) */
  content: string;
}

export interface TikZOption {
  key: string;
  /** Empty string for flags like "thick" */
  value: string;
}
