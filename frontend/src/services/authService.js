import axios from 'axios';

const API_URL = import.meta.env.VITE_USERS_API_URL;

const authService = {
  // Login
  login: async (login, password) => {
    const response = await axios.post(`${API_URL}/users/login`, {
      login,
      password,
    });
    return response.data;
  },

  // Register
  register: async (userData) => {
    const response = await axios.post(`${API_URL}/users/register`, userData);
    return response.data;
  },

  // Get user by ID
  getUserById: async (userId, token) => {
    const response = await axios.get(`${API_URL}/users/${userId}`, {
      headers: {
        Authorization: `Bearer ${token}`,
      },
    });
    return response.data;
  },
};

export default authService;