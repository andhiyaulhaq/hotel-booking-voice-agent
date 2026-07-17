import React, { useEffect, useState } from 'react';

export const CheckoutModal: React.FC = () => {
  const [isOpen, setIsOpen] = useState(false);
  const [url, setUrl] = useState('');

  useEffect(() => {
    const handleShowCheckout = (e: Event) => {
      const customEvent = e as CustomEvent;
      setUrl(customEvent.detail);
      setIsOpen(true);
    };

    window.addEventListener('show_checkout', handleShowCheckout);
    return () => window.removeEventListener('show_checkout', handleShowCheckout);
  }, []);

  if (!isOpen) return null;

  return (
    <div className="modal-backdrop">
      <div className="modal-content glass">
        <h2>Complete Your Booking</h2>
        <p>Please complete your payment to finalize the reservation.</p>
        {url && (
            <iframe src={url} width="100%" height="500px" style={{border: 'none', borderRadius: '8px', marginTop: '1rem'}} />
        )}
        <button onClick={() => setIsOpen(false)} className="close-btn" style={{marginTop: '1rem', width: '100%'}}>Close</button>
      </div>
    </div>
  );
};
