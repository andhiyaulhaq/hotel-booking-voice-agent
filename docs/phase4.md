# Phase 4: The Interfaces (Web UI)

## Objective
Build the visual interfaces that bring the Voice AI to life. We will build a Guest Booking Portal and an Admin Dashboard using modern web technologies (HTML/JS/CSS or Vite).

---

## 1. Project Structure
This phase creates the UI directories:
```text
ui/
├── guest/
│   ├── index.html            # Main booking portal layout
│   ├── css/
│   │   └── style.css         # Glassmorphism luxury styles
│   └── js/
│       ├── audio.js          # Microphone capture & WebAudio API
│       ├── socket.js         # WebSocket connection to Go Gateway
│       └── visualizer.js     # HTML5 Canvas waveforms
└── admin/
    ├── index.html            # Dashboard layout
    ├── css/
    │   └── admin.css         
    └── js/
        └── dashboard.js      # WebSocket/SSE logic for real-time ledger
```

## 2. Project Initialization
```bash
mkdir -p ui/guest/css ui/guest/js ui/admin/css ui/admin/js
```

## 3. Guest UI (The AI Booking Portal)
Implement a modern, glassmorphism-styled hotel website.

### Visual Layout
- **Background:** High-quality imagery of luxury suites.
- **Main Content:** Traditional date pickers, room cards, and prices.
- **Voice Overlay (Bottom Right):** A floating action button with a microphone icon.

### Voice Interaction Logic
- Use `navigator.mediaDevices.getUserMedia` to capture raw microphone audio.
- Resample audio to 16kHz PCM if required by Cartesia.
- Connect to the Go Edge Gateway (`ws://localhost:8080/ws`).
- **Dynamic Visualizer:** Draw audio waveforms on an HTML5 `<canvas>` based on the microphone volume and incoming TTS volume.

### UI Sync (Multimodal)
- Listen for specific JSON events from the WebSocket.
- If `type: "show_checkout"` arrives, visually dim the screen and pop up a modal containing the Xendit QRIS code or Invoice iframe.

## 4. Admin UI (The Manager Dashboard)
Implement a data-heavy back-office view.

### Visual Layout
- **Dashboard Grid:** Show total rooms, occupied rooms, and pending checkouts.
- **Live Ledger:** A scrolling table showing recent bookings (`ID, Guest Name, Room, Status`).

### Real-Time Updates
- Connect to the Go Edge Gateway via a separate admin WebSocket or Server-Sent Events (SSE).
- When Xendit confirms a payment, Go broadcasts an update to the Admin UI, causing the ledger to instantly change from `pending_payment` to `confirmed` with a green checkmark, demonstrating real-time data sync without manual page refreshes.

---

## 5. Test Scenarios

### Manual Verification (Localhost)
1. Serve the `ui/guest` and `ui/admin` folders using a simple HTTP server (e.g., `python -m http.server`).
2. **Audio Test:** Click the microphone button and ensure browser permissions prompt correctly. Speak into the mic and ensure the `audio_in` WebSocket messages are firing.
3. **Multimodal Test:** Hardcode a mock `show_checkout` WebSocket message in the Go server and verify the UI dims and pops up the QRIS modal correctly.
4. **Admin Dashboard Test:** Open the Admin UI in a separate window. Trigger a booking in the Guest UI and verify the Admin UI table adds a row instantly without needing a page refresh.
