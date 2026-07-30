import React, { useEffect, useState } from 'react';

interface Booking {
  id: number;
  guest_name: string;
  room_type: string;
  nights: number;
  status: string;
}

interface RoomInventory {
  room_type: string;
  available: number;
  total: number;
}

export const AdminDashboard: React.FC = () => {
  const [bookings, setBookings] = useState<Booking[]>([]);
  const [inventory, setInventory] = useState<RoomInventory[]>([]);
  const [loading, setLoading] = useState(true);

  const fetchData = async () => {
    try {
      const [bookingsRes, inventoryRes] = await Promise.all([
        fetch('http://localhost:8080/api/bookings'),
        fetch('http://localhost:8080/api/inventory')
      ]);

      if (bookingsRes.ok) {
        const data = await bookingsRes.json();
        setBookings(data);
      }
      if (inventoryRes.ok) {
        const data = await inventoryRes.json();
        setInventory(data);
      }
    } catch (err) {
      console.error("Failed to fetch dashboard data:", err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    // Initial fetch
    fetchData();

    // Poll every 5 seconds
    const interval = setInterval(fetchData, 5000);
    return () => clearInterval(interval);
  }, []);

  return (
    <div className="admin-container">
      <header className="glass-header" style={{ marginBottom: '2rem' }}>
        <h1>Admin Dashboard</h1>
      </header>
      <main className="admin-main">
        {/* Live Inventory Widget */}
        <div className="glass" style={{ padding: '2rem', marginBottom: '2rem' }}>
          <h2>Live Inventory</h2>
          <div style={{ display: 'flex', gap: '1rem', flexWrap: 'wrap', marginTop: '1rem' }}>
            {inventory.length === 0 && loading ? (
              <p>Loading inventory...</p>
            ) : (
              inventory.map((inv) => (
                <div key={inv.room_type} style={{
                  flex: '1 1 200px',
                  padding: '1.5rem',
                  borderRadius: '12px',
                  background: 'rgba(255,255,255,0.05)',
                  border: '1px solid rgba(255,255,255,0.1)',
                  textAlign: 'center'
                }}>
                  <h3 style={{ textTransform: 'capitalize', margin: '0 0 1rem 0' }}>{inv.room_type}</h3>
                  <div style={{ fontSize: '2rem', fontWeight: 'bold' }}>
                    <span style={{ color: inv.available > 0 ? '#4caf50' : '#f44336' }}>{inv.available}</span>
                    <span style={{ color: 'rgba(255,255,255,0.5)', fontSize: '1.5rem' }}> / {inv.total}</span>
                  </div>
                  <div style={{ marginTop: '0.5rem', fontSize: '0.9rem', color: 'rgba(255,255,255,0.7)' }}>
                    Available
                  </div>
                </div>
              ))
            )}
          </div>
        </div>

        {/* Live Booking Ledger */}
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
