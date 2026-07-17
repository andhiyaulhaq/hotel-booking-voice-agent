import React from 'react';

export const AdminDashboard: React.FC = () => {
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
                    <tr>
                        <td>1</td>
                        <td>Alice</td>
                        <td>Suite</td>
                        <td><span className="status pending">Pending Payment</span></td>
                    </tr>
                </tbody>
            </table>
        </div>
      </main>
    </div>
  );
};
