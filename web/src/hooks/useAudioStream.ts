import { useState, useRef, useCallback } from 'react';

export function useAudioStream(onAudioData: (base64Audio: string) => void) {
  const [isRecording, setIsRecording] = useState(false);
  const audioContextRef = useRef<AudioContext | null>(null);
  const analyserRef = useRef<AnalyserNode | null>(null);
  const nextPlayTimeRef = useRef<number>(0);
  const [volume, setVolume] = useState(0);

  const startRecording = useCallback(async () => {
    try {
      const stream = await navigator.mediaDevices.getUserMedia({ 
        audio: {
          echoCancellation: true,
          noiseSuppression: true,
          autoGainControl: true
        } 
      });
      audioContextRef.current = new (window.AudioContext || (window as any).webkitAudioContext)({ sampleRate: 16000 });
      
      analyserRef.current = audioContextRef.current.createAnalyser();
      const source = audioContextRef.current.createMediaStreamSource(stream);
      source.connect(analyserRef.current);

      // We use a ScriptProcessorNode to get raw PCM data (AudioWorklet is better for prod, but this is simpler for the demo)
      const processor = audioContextRef.current.createScriptProcessor(4096, 1, 1);
      processor.onaudioprocess = (e) => {
        // Software Mute: If the AI is currently speaking, ignore the microphone to prevent echo loops
        // This acts as a foolproof half-duplex echo cancellation.
        if (audioContextRef.current && nextPlayTimeRef.current > audioContextRef.current.currentTime) {
            return;
        }

        const inputData = e.inputBuffer.getChannelData(0);
        
        // Convert Float32 to Int16 PCM
        const pcm16 = new Int16Array(inputData.length);
        for (let i = 0; i < inputData.length; i++) {
          pcm16[i] = Math.max(-1, Math.min(1, inputData[i])) * 32767;
        }
        
        // Base64 encode the binary data
        const bytes = new Uint8Array(pcm16.buffer);
        let binary = '';
        for (let i = 0; i < bytes.byteLength; i++) {
            binary += String.fromCharCode(bytes[i]);
        }
        const base64 = btoa(binary);
        onAudioData(base64);

        // Update volume state for visualizer
        if (analyserRef.current) {
            const dataArray = new Uint8Array(analyserRef.current.frequencyBinCount);
            analyserRef.current.getByteFrequencyData(dataArray);
            let sum = 0;
            for (let i = 0; i < dataArray.length; i++) {
                sum += dataArray[i];
            }
            setVolume(sum / dataArray.length);
        }
      };

      source.connect(processor);
      // Removed: processor.connect(audioContextRef.current.destination);
      // We don't want the user's mic to play back into their own earphones (sidetone)!
      // The ScriptProcessorNode still fires onaudioprocess even if not connected to destination in most browsers,
      // but to be safe we can connect it to a dummy gain node with 0 volume.
      const dummyGain = audioContextRef.current.createGain();
      dummyGain.gain.value = 0;
      processor.connect(dummyGain);
      dummyGain.connect(audioContextRef.current.destination);
      
      setIsRecording(true);
    } catch (err) {
      console.error("Error accessing microphone", err);
    }
  }, [onAudioData]);

  const stopRecording = useCallback(() => {
    if (audioContextRef.current) {
      audioContextRef.current.close();
      audioContextRef.current = null;
    }
    setIsRecording(false);
    setVolume(0);
  }, []);

  const playAudio = useCallback((base64PCM: string) => {
    if (!audioContextRef.current) {
      audioContextRef.current = new (window.AudioContext || (window as any).webkitAudioContext)({ sampleRate: 16000 });
      nextPlayTimeRef.current = audioContextRef.current.currentTime;
    }
    const binary = atob(base64PCM);
    const len = binary.length;
    const bytes = new Uint8Array(len);
    for (let i = 0; i < len; i++) {
        bytes[i] = binary.charCodeAt(i);
    }
    
    // PCM Int16 to Float32
    const pcm16 = new Int16Array(bytes.buffer);
    const audioBuffer = audioContextRef.current.createBuffer(1, pcm16.length, 16000);
    const channelData = audioBuffer.getChannelData(0);
    for (let i = 0; i < pcm16.length; i++) {
      channelData[i] = pcm16[i] / 32768.0;
    }
    
    const source = audioContextRef.current.createBufferSource();
    source.buffer = audioBuffer;
    source.connect(audioContextRef.current.destination);
    
    // Schedule playback sequentially to avoid overlapping chunks
    const currentTime = audioContextRef.current.currentTime;
    if (nextPlayTimeRef.current < currentTime) {
        nextPlayTimeRef.current = currentTime;
    }
    
    source.start(nextPlayTimeRef.current);
    nextPlayTimeRef.current += audioBuffer.duration;
  }, []);

  return { isRecording, startRecording, stopRecording, playAudio, volume };
}
