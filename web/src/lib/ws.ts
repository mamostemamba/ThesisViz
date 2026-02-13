export interface WSMessage {
  type: "status" | "preview" | "result" | "error";
  phase:
    | "planning"
    | "generating"
    | "compiling"
    | "reviewing"
    | "fixing"
    | "rerolling"
    | "done";
  /** Client-side timestamp (ms) set when the message is received. */
  _ts?: number;
  data: {
    message?: string;
    progress?: number;
    round?: number;
    image_url?: string;
    issues?: string[];
    // result fields
    generation_id?: string;
    code?: string;
    format?: string;
    explanation?: string;
    review_passed?: boolean;
    review_rounds?: number;
    critique?: string;
    score?: number;
    full_tex?: string;
  };
}

const WS_BASE =
  process.env.NEXT_PUBLIC_WS_URL || "ws://localhost:8080";

/**
 * Connect to the generation progress WebSocket.
 * Returns a cleanup function to close the connection.
 */
export function connectGeneration(
  taskId: string,
  onMessage: (msg: WSMessage) => void,
  onClose?: () => void
): () => void {
  const url = `${WS_BASE}/api/v1/ws/generate/${taskId}`;
  const ws = new WebSocket(url);

  ws.onmessage = (event) => {
    try {
      const msg: WSMessage = JSON.parse(event.data);
      msg._ts = Date.now();
      onMessage(msg);
    } catch {
      // ignore malformed messages
    }
  };

  ws.onclose = () => {
    onClose?.();
  };

  ws.onerror = () => {
    ws.close();
  };

  return () => {
    ws.close();
  };
}
