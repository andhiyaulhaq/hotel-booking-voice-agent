import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { GuestPortal } from './pages/GuestPortal';
import { AdminDashboard } from './pages/AdminDashboard';
import './index.css';

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<GuestPortal />} />
        <Route path="/admin" element={<AdminDashboard />} />
      </Routes>
    </BrowserRouter>
  );
}

export default App;
