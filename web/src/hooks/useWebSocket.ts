import { useState, useEffect, useRef, useCallback } from 'react';

export interface WebSocketMessage {
  type: string;
  data: string;
  url?: string;
}

export function useWebSocket(url: string, onMessage?: (msg: WebSocketMessage) => void) {
  const [isConnected, setIsConnected] = useState(false);
  const [messages, setMessages] = useState<WebSocketMessage[]>([]);
  const wsRef = useRef<WebSocket | null>(null);

  useEffect(() => {
    const ws = new WebSocket(url);
    wsRef.current = ws;

    ws.onopen = () => {
      console.log('Connected to WS');
      setIsConnected(true);
    };

    ws.onmessage = (event) => {
      try {
        const msg: WebSocketMessage = JSON.parse(event.data);
        setMessages((prev) => [...prev, msg]);
        if (onMessage) {
            onMessage(msg);
        }
      } catch (e) {
        console.error('Failed to parse WS message', e);
      }
    };

    ws.onclose = () => {
      console.log('Disconnected from WS');
      setIsConnected(false);
    };

    return () => {
      ws.close();
    };
  }, [url]);

  const sendMessage = useCallback((msg: WebSocketMessage) => {
    if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify(msg));
    }
  }, []);

  const clearMessages = useCallback(() => {
    setMessages([]);
  }, []);

  return { isConnected, messages, sendMessage, clearMessages };
}
