import React from 'react';
import { VoiceOverlay } from '../components/VoiceOverlay';
import { CheckoutModal } from '../components/CheckoutModal';

export const GuestPortal: React.FC = () => {
  return (
    <div className="portal-container">
      <header className="glass-header">
        <h1>The Grand AI Hotel</h1>
        <nav>
          <span>Rooms</span>
          <span>Dining</span>
          <span>Experiences</span>
        </nav>
      </header>
      
      <main className="hero-section">
        <div className="hero-content glass">
          <h2>Experience Luxury Redefined</h2>
          <p>Talk to our Voice Concierge to book your stay instantly.</p>
        </div>
      </main>

      <VoiceOverlay />
      <CheckoutModal />
    </div>
  );
};
