import React, { useEffect, useState } from 'react';

interface Booking {
  id: number;
  guest_name: string;
  room_type: string;
  nights: number;
  status: string;
}

export const AdminDashboard: React.FC = () => {
  const [bookings, setBookings] = useState<Booking[]>([]);
  const [loading, setLoading] = useState(true);

  const fetchBookings = async () => {
    try {
      const response = await fetch('http://localhost:8080/api/bookings');
      if (response.ok) {
        const data = await response.json();
        setBookings(data);
      }
    } catch (err) {
      console.error("Failed to fetch bookings:", err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    // Initial fetch
    fetchBookings();

    // Poll every 5 seconds
    const interval = setInterval(fetchBookings, 5000);
    return () => clearInterval(interval);
  }, []);

  return (
    <div className="admin-container">
      <header className="glass-header" style={{ marginBottom: '2rem' }}>
        <h1>Admin Dashboard</h1>
      </header>
      <main className="admin-main">
        <div className="glass" style={{ padding: '2rem' }}>
            <h2>Live Booking Ledger</h2>
            <table>
                <thead>
                    <tr>
                        <th>ID</th>
                        <th>Guest Name</th>
                        <th>Room Type</th>
                        <th>Status</th>
                    </tr>
                </thead>
                <tbody>
                    {loading && bookings.length === 0 ? (
                        <tr><td colSpan={4} style={{ textAlign: 'center' }}>Loading ledger...</td></tr>
                    ) : bookings.length === 0 ? (
                        <tr><td colSpan={4} style={{ textAlign: 'center' }}>No bookings found.</td></tr>
                    ) : (
                        bookings.map((booking) => (
                            <tr key={booking.id}>
                                <td>{booking.id}</td>
                                <td>{booking.guest_name}</td>
                                <td style={{ textTransform: 'capitalize' }}>{booking.room_type} ({booking.nights} nights)</td>
                                <td>
                                    <span className={`status ${booking.status === 'confirmed' ? 'confirmed' : 'pending'}`}>
                                        {booking.status === 'pending_payment' ? 'Pending Payment' : booking.status}
                                    </span>
                                </td>
                            </tr>
                        ))
                    )}
                </tbody>
            </table>
        </div>
      </main>
    </div>
  );
};
