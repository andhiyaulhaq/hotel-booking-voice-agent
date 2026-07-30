# ADR 0001: Use AudioWorklet for Real-Time Audio Processing

## Status
Accepted

## Context
Our multimodal web frontend (`web/src/hooks/useAudioStream.ts`) was originally built using the `ScriptProcessorNode` to capture raw PCM audio from the user's microphone and convert it to Base64 chunks for WebSocket streaming.

During development, the IDE flagged `ScriptProcessorNode` and its associated `onaudioprocess` event handler as deprecated. 

The `ScriptProcessorNode` has a fundamental architectural flaw: it executes entirely on the main UI thread. In a modern React application, especially one that renders dynamic visualizers or complex glassmorphism UI components, the main thread can experience heavy load. If the main thread is blocked by rendering tasks, the `onaudioprocess` callback will be delayed, resulting in audio buffer under-runs, dropped frames, stuttering, and increased latency. In a real-time Voice AI agent where latency is critical, this is unacceptable.

The Web Audio API introduced `AudioWorkletNode` as the modern replacement.

## Decision
We will migrate our audio processing pipeline from `ScriptProcessorNode` to `AudioWorkletNode`.

1. **Dedicated Processing Thread**: We will create an independent JavaScript file (`audio-processor.js`) placed in the `public` directory.
2. **AudioWorkletProcessor**: The script will implement the `AudioWorkletProcessor` interface to handle the Float32 to Int16 PCM conversion.
3. **Asynchronous Messaging**: We will use the `port.postMessage` API to pass the processed audio chunks back to the main React thread asynchronously, ensuring the main thread is never blocked by raw audio computation.
4. **Half-Duplex Muting**: We will pass state messages to the worklet (e.g., `isPlaying: true`) to maintain our software-based echo cancellation (dropping mic frames while the AI is speaking) directly on the background thread.

## Consequences
### Positive
- **Zero-Latency Processing**: Audio processing is offloaded to a dedicated audio thread, completely immunizing the voice stream from React render cycle jitters.
- **Future-Proof**: Eliminates IDE deprecation warnings and ensures compatibility with future browser releases that may remove `ScriptProcessorNode` entirely.
- **Cleaner Main Thread**: Reduces the computational burden on the main UI thread.

### Negative
- **Architectural Complexity**: Introduces an asynchronous, multi-threaded architecture to the frontend.
- **Build/Deployment Considerations**: The worklet must be loaded as a separate file via HTTP (`await audioContext.audioWorklet.addModule('/audio-processor.js')`), which requires ensuring the file is properly served from the `public` directory and bypasses bundler obfuscation.
