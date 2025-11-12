import axios from 'axios';

const API_URL = import.meta.env.VITE_CANCHAS_API_URL;

const canchaService = {
  // Obtener todas las canchas
  getAllCanchas: async () => {
    const response = await axios.get(`${API_URL}/canchas`);
    return response.data;
  },

  // Obtener cancha por ID
  getCanchaById: async (id) => {
    const response = await axios.get(`${API_URL}/canchas/${id}`);
    return response.data;
  },

  // Crear cancha (solo admin)
  createCancha: async (canchaData, token) => {
    const response = await axios.post(`${API_URL}/canchas`, canchaData, {
      headers: {
        Authorization: `Bearer ${token}`,
      },
    });
    return response.data;
  },

  // Actualizar cancha (solo admin)
  updateCancha: async (id, canchaData, token) => {
    const response = await axios.put(`${API_URL}/canchas/${id}`, canchaData, {
      headers: {
        Authorization: `Bearer ${token}`,
      },
    });
    return response.data;
  },

  // Eliminar cancha (solo admin)
  deleteCancha: async (id, token) => {
    const response = await axios.delete(`${API_URL}/canchas/${id}`, {
      headers: {
        Authorization: `Bearer ${token}`,
      },
    });
    return response.data;
  },
};

export default canchaService;