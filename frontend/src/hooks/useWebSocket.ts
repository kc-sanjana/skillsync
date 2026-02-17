// src/hooks/useWebSocket.ts
import { useState, useEffect, useRef, useCallback } from "react";
import { Message } from "../types";

/**
 * ✅ Vite-compatible environment variable
 */
const WS_BASE_URL =
  import.meta.env.VITE_WS_BASE_URL || "ws://localhost:8080";

interface WebSocketMessagePayload {
  type: string;
  data: any;
}

interface UseWebSocketProps {
  token: string | null;
  matchId: string | null;
}

const useWebSocket = ({ token, matchId }: UseWebSocketProps) => {
  const [messages, setMessages] = useState<Message[]>([]);
  const [isConnected, setIsConnected] = useState(false);
  const [typingUsers, setTypingUsers] = useState<string[]>([]);

  const ws = useRef<WebSocket | null>(null);
  const reconnectTimeout = useRef<number | null>(null);
  const reconnectAttempts = useRef(0);

  const connect = useCallback(() => {
    if (!token || !matchId) {
      console.warn("WebSocket: Missing token or matchId");
      setIsConnected(false);
      return;
    }

    if (ws.current && ws.current.readyState === WebSocket.OPEN) return;

    const url = `${WS_BASE_URL}/ws?token=${token}&match_id=${matchId}`;
    ws.current = new WebSocket(url);

    ws.current.onopen = () => {
      console.log("WebSocket Connected");
      setIsConnected(true);
      reconnectAttempts.current = 0;

      if (reconnectTimeout.current) {
        clearTimeout(reconnectTimeout.current);
        reconnectTimeout.current = null;
      }
    };

    ws.current.onmessage = (event) => {
      const parsed: WebSocketMessagePayload = JSON.parse(event.data);

      switch (parsed.type) {
        case "message":
          setMessages((prev) => [...prev, parsed.data as Message]);
          break;

        case "typing":
          setTypingUsers((prev) => {
            const username = parsed.data.username;
            return parsed.data.isTyping
              ? prev.includes(username)
                ? prev
                : [...prev, username]
              : prev.filter((u) => u !== username);
          });
          break;

        case "code_change":
          console.log("Code change:", parsed.data.code);
          break;

        default:
          console.warn("Unknown WS message:", parsed.type);
      }
    };

    ws.current.onclose = () => {
      console.log("WebSocket Disconnected");
      setIsConnected(false);

      reconnectAttempts.current += 1;
      const delay = Math.min(1000 * 2 ** reconnectAttempts.current, 30000);

      reconnectTimeout.current = window.setTimeout(connect, delay);
    };

    ws.current.onerror = (error) => {
      console.error("WebSocket Error:", error);
      ws.current?.close();
    };
  }, [token, matchId]);

  useEffect(() => {
    connect();

    return () => {
      ws.current?.close();
      if (reconnectTimeout.current) clearTimeout(reconnectTimeout.current);
    };
  }, [connect]);

  const sendWebSocketMessage = useCallback((type: string, data: any) => {
    if (ws.current?.readyState === WebSocket.OPEN) {
      ws.current.send(JSON.stringify({ type, data }));
    } else {
      console.warn("WebSocket not connected");
    }
  }, []);

  const sendMessage = useCallback(
    (content: string) => sendWebSocketMessage("message", { content }),
    [sendWebSocketMessage]
  );

  const sendTyping = useCallback(
    (isTyping: boolean) => sendWebSocketMessage("typing", { isTyping }),
    [sendWebSocketMessage]
  );

  const sendCodeChange = useCallback(
    (code: string) => sendWebSocketMessage("code_change", { code }),
    [sendWebSocketMessage]
  );

  return {
    sendMessage,
    sendTyping,
    sendCodeChange,
    messages,
    isConnected,
    typingUsers,
  };
};

export default useWebSocket;

