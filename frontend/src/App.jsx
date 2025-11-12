import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider } from './context/AuthContext';
import Navbar from './components/Navbar';
import PrivateRoute from './components/PrivateRoute';
import AdminRoute from './components/AdminRoute';

// Pages
import Login from './pages/Login';
import Register from './pages/Register';
import Home from './pages/Home';
import CanchaDetails from './pages/CanchaDetails';
import Congrats from './pages/Congrats';
import MisReservas from './pages/MisReservas';
import Admin from './pages/Admin';

import './App.css';

function App() {
  return (
    <AuthProvider>
      <BrowserRouter>
        <div className="App">
          <Navbar />
          <Routes>
            {/* Rutas públicas */}
            <Route path="/login" element={<Login />} />
            <Route path="/register" element={<Register />} />
            <Route path="/" element={<Home />} />
            <Route path="/cancha/:id" element={<CanchaDetails />} />

            {/* Rutas protegidas (requieren autenticación) */}
            <Route
              path="/congrats"
              element={
                <PrivateRoute>
                  <Congrats />
                </PrivateRoute>
              }
            />
            <Route
              path="/mis-reservas"
              element={
                <PrivateRoute>
                  <MisReservas />
                </PrivateRoute>
              }
            />

            {/* Rutas de admin (requieren rol admin) */}
            <Route
              path="/admin"
              element={
                <AdminRoute>
                  <Admin />
                </AdminRoute>
              }
            />

            {/* Ruta por defecto */}
            <Route path="*" element={<Navigate to="/" replace />} />
          </Routes>
        </div>
      </BrowserRouter>
    </AuthProvider>
  );
}

export default App;