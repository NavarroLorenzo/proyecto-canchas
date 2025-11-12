import axios from 'axios';

const API_URL = import.meta.env.VITE_API_URL;

const searchService = {
  // Buscar canchas
  searchCanchas: async (params) => {
    const response = await axios.get(`${API_URL}/search`, { params });
    return response.data;
  },
};

export default searchService;