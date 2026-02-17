// src/hooks/useWebSocket.ts
import { useState, useEffect, useRef, useCallback } from "react";
import { Message } from "../types";

/**
 * Vite-compatible environment variable
 */
const WS_BASE_URL =
  import.meta.env.VITE_WS_BASE_URL || "ws://localhost:8080";

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
      try {
        const parsed = JSON.parse(event.data);

        switch (parsed.type) {
          case "chat_message":
            if (parsed.message) {
              const msg: Message = {
                id: parsed.message.id?.toString() || "",
                match_id: parsed.message.match_id?.toString() || matchId || "",
                sender_id: parsed.message.sender_id || "",
                content: parsed.message.content || "",
                timestamp: parsed.message.created_at || parsed.timestamp || new Date().toISOString(),
              };
              setMessages((prev) => [...prev, msg]);
            }
            break;

          case "typing_indicator":
            setTypingUsers((prev) => {
              const userId = parsed.user_id;
              return parsed.is_typing
                ? prev.includes(userId)
                  ? prev
                  : [...prev, userId]
                : prev.filter((u) => u !== userId);
            });
            break;

          case "code_change":
            console.log("Code change:", parsed.code);
            break;

          default:
            console.warn("Unknown WS message:", parsed.type);
        }
      } catch (err) {
        console.error("Failed to parse WS message:", err);
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
    (content: string) => sendWebSocketMessage("chat_message", { content }),
    [sendWebSocketMessage]
  );

  const sendTyping = useCallback(
    (isTyping: boolean) => sendWebSocketMessage("typing_indicator", { is_typing: isTyping }),
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
