import axios from 'axios';

// Configurar interceptor de respuesta para detectar tokens expirados
axios.interceptors.response.use(
  (response) => {
    return response;
  },
  (error) => {
    // Si recibimos un 401 (Unauthorized), el token probablemente expiró
    if (error.response && error.response.status === 401) {
      // Limpiar el localStorage
      localStorage.removeItem('token');
      localStorage.removeItem('user');
      
      // Redirigir al login si no estamos ya ahí
      if (window.location.pathname !== '/login' && window.location.pathname !== '/register') {
        alert('Tu sesión ha expirado. Por favor, inicia sesión nuevamente.');
        window.location.href = '/login';
      }
    }
    return Promise.reject(error);
  }
);

export default axios;

