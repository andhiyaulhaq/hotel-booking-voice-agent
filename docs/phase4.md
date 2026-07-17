# Phase 4: The Interfaces (Web UI)

## Objective
Build the visual interfaces that bring the Voice AI to life. We will build a Guest Booking Portal and an Admin Dashboard using **Vite + React**.

---

## 1. Project Structure
This phase creates the React frontend in the `web/` directory:
```text
web/
├── index.html
├── src/
│   ├── main.tsx              # Vite + React entry point
│   ├── App.tsx               # React Router (Guest vs Admin routes)
│   ├── pages/
│   │   ├── GuestPortal.tsx   # Main booking UI
│   │   └── AdminDashboard.tsx# Real-time ledger
│   ├── components/
│   │   ├── VoiceOverlay.tsx  # Floating mic & visualizer
│   │   └── CheckoutModal.tsx # Xendit QRIS modal
│   ├── hooks/
│   │   ├── useAudioStream.ts # WebAudio API logic
│   │   └── useWebSocket.ts   # Go Gateway connection
│   └── styles/
│       └── index.css         # Vanilla CSS Glassmorphism tokens
```

## 2. Project Initialization
```bash
# From the project root
pnpm create vite web --template react-ts
cd web
pnpm add react-router-dom
```

## 3. Guest UI (The AI Booking Portal)
Implement a modern, glassmorphism-styled hotel website using React components and Vanilla CSS.

### Visual Layout
- **Background:** High-quality imagery of luxury suites.
- **Main Content:** Traditional date pickers, room cards, and prices.
- **Voice Overlay (Bottom Right):** A floating action button with a microphone icon managed by the `<VoiceOverlay />` component.

### Voice Interaction Logic (Custom Hooks)
- Use `useAudioStream` hook to call `navigator.mediaDevices.getUserMedia` to capture raw microphone audio.
- Use `useWebSocket` to connect to the Go Edge Gateway (`ws://localhost:8080/ws`).
- **Dynamic Visualizer:** Draw audio waveforms on an HTML5 `<canvas>` based on the microphone volume and incoming TTS volume.

### UI Sync (Multimodal)
- The React state listens for specific JSON events from the WebSocket.
- If `type: "show_checkout"` arrives, visually dim the screen and mount the `<CheckoutModal />` containing the Xendit QRIS code or Invoice iframe.

## 4. Admin UI (The Manager Dashboard)
Implement a data-heavy back-office view.

### Visual Layout
- **Dashboard Grid:** Show total rooms, occupied rooms, and pending checkouts.
- **Live Ledger:** A scrolling table showing recent bookings (`ID, Guest Name, Room, Status`).

### Real-Time Updates
- Connect to the Go Edge Gateway via a separate admin WebSocket or Server-Sent Events (SSE).
- When Xendit confirms a payment, Go broadcasts an update to the Admin UI, causing the ledger's React state to instantly change from `pending_payment` to `confirmed` with a green checkmark, demonstrating real-time data sync without manual page refreshes.

---

## 5. Test Scenarios

### Manual Verification (Localhost)
1. Run the Vite dev server `pnpm run dev` in the `web/` directory.
2. **Audio Test:** Click the microphone button and ensure browser permissions prompt correctly. Speak into the mic and ensure the `audio_in` WebSocket messages are firing.
3. **Multimodal Test:** Hardcode a mock `show_checkout` WebSocket message in the Go server and verify the UI dims and pops up the QRIS modal correctly.
4. **Admin Dashboard Test:** Open the Admin Dashboard route (`/admin`) in a separate window. Trigger a booking in the Guest Portal and verify the Admin table adds a row instantly.
