import React, { useState, useCallback, useRef, useEffect } from 'react';
import { useAudioStream } from '../hooks/useAudioStream';
import { useWebSocket } from '../hooks/useWebSocket';

export const VoiceOverlay: React.FC = () => {
  const [isActive, setIsActive] = useState(false);
  
  const playAudioRef = useRef<((data: string) => void) | null>(null);

  const handleWebSocketMessage = useCallback((msg: any) => {
    if (msg.type === 'audio_out') {
      if (playAudioRef.current) {
        playAudioRef.current(msg.data);
      }
    } else if (msg.type === 'show_checkout') {
      window.dispatchEvent(new CustomEvent('show_checkout', { detail: msg.url }));
    }
  }, []);

  const { isConnected, sendMessage } = useWebSocket('ws://localhost:8080/ws', handleWebSocketMessage);
  
  const handleAudioData = useCallback((base64Audio: string) => {
    sendMessage({ type: 'audio_in', data: base64Audio });
  }, [sendMessage]);
  
  const { isRecording, startRecording, stopRecording, playAudio, volume } = useAudioStream(handleAudioData);

  useEffect(() => {
    playAudioRef.current = playAudio;
  }, [playAudio]);

  const toggleRecording = () => {
    if (isRecording) {
      stopRecording();
      setIsActive(false);
    } else {
      startRecording();
      setIsActive(true);
    }
  };

  return (
    <div className={`voice-overlay ${isActive ? 'active' : ''} glass`}>
      <div className="status-indicator">
        {isConnected ? <span className="dot green" /> : <span className="dot red" />}
      </div>
      <button className={`mic-button ${isActive ? 'recording' : ''}`} onClick={toggleRecording}>
        {isActive ? '⏹ Stop' : '🎤 Speak to Concierge'}
      </button>
      {isActive && (
        <div className="visualizer-container">
           <div className="visualizer-bar" style={{ height: `${Math.max(4, volume / 2)}px` }} />
        </div>
      )}
    </div>
  );
};
