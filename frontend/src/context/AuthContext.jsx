import { createContext, useContext, useState, useEffect, useRef } from 'react';
import { isTokenExpired } from '../utils/jwt';

const AuthContext = createContext();

export const AuthProvider = ({ children }) => {
  const [user, setUser] = useState(null);
  const [token, setToken] = useState(null);
  const [loading, setLoading] = useState(true);
  const expirationCheckInterval = useRef(null);
  const expirationTimeout = useRef(null);

  // Función para desloguear y mostrar aviso cuando el token expira
  const handleTokenExpiration = () => {
    setUser(null);
    setToken(null);
    localStorage.removeItem('token');
    localStorage.removeItem('user');
    alert('Tu sesión ha expirado. Por favor, inicia sesión nuevamente.');
    window.location.href = '/login';
  };

  // Verificar expiración del token periódicamente
  useEffect(() => {
    const checkTokenExpiration = () => {
      if (!token) return;

      // Solo verificar si el token está expirado, sin avisos previos
      if (isTokenExpired(token)) {
        // Token expirado, desloguear y mostrar aviso
        handleTokenExpiration();
        return;
      }
    };

    // Verificar cada 10 segundos
    if (token) {
      expirationCheckInterval.current = setInterval(checkTokenExpiration, 10000);
      // Verificar inmediatamente
      checkTokenExpiration();
    }

    return () => {
      if (expirationCheckInterval.current) {
        clearInterval(expirationCheckInterval.current);
      }
      if (expirationTimeout.current) {
        clearTimeout(expirationTimeout.current);
      }
    };
  }, [token]);

  useEffect(() => {
    // Cargar usuario y token del localStorage al iniciar
    const storedToken = localStorage.getItem('token');
    const storedUser = localStorage.getItem('user');

    if (storedToken && storedUser) {
      // Verificar si el token está expirado al cargar
      if (isTokenExpired(storedToken)) {
        localStorage.removeItem('token');
        localStorage.removeItem('user');
        setLoading(false);
        return;
      }
      
      setToken(storedToken);
      setUser(JSON.parse(storedUser));
    }
    setLoading(false);
  }, []);

  const login = (userData, authToken) => {
    setUser(userData);
    setToken(authToken);
    localStorage.setItem('token', authToken);
    localStorage.setItem('user', JSON.stringify(userData));
  };

  const logout = () => {
    setUser(null);
    setToken(null);
    localStorage.removeItem('token');
    localStorage.removeItem('user');
  };

  const isAdmin = () => {
    return user?.role === 'admin';
  };

  const value = {
    user,
    token,
    login,
    logout,
    isAdmin,
    isAuthenticated: !!token && !isTokenExpired(token),
  };

  if (loading) {
    return <div>Loading...</div>;
  }

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
};

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};