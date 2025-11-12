import axios from 'axios';

const API_URL = import.meta.env.VITE_RESERVAS_API_URL;

const reservaService = {
  // Crear reserva
  createReserva: async (reservaData, token) => {
    const response = await axios.post(`${API_URL}/reservas`, reservaData, {
      headers: {
        Authorization: `Bearer ${token}`,
      },
    });
    return response.data;
  },

  // Obtener todas las reservas
  getAllReservas: async (token) => {
    const response = await axios.get(`${API_URL}/reservas`, {
      headers: {
        Authorization: `Bearer ${token}`,
      },
    });
    return response.data;
  },

  // Obtener reservas de un usuario
  getReservasByUserId: async (userId, token) => {
    const response = await axios.get(`${API_URL}/reservas/user/${userId}`, {
      headers: {
        Authorization: `Bearer ${token}`,
      },
    });
    return response.data;
  },

  // Obtener reserva por ID
  getReservaById: async (id, token) => {
    const response = await axios.get(`${API_URL}/reservas/${id}`, {
      headers: {
        Authorization: `Bearer ${token}`,
      },
    });
    return response.data;
  },

  // Cancelar reserva
  cancelReserva: async (id, token) => {
    const response = await axios.delete(`${API_URL}/reservas/${id}`, {
      headers: {
        Authorization: `Bearer ${token}`,
      },
    });
    return response.data;
  },
};

export default reservaService;