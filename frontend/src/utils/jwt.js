// Utilidad para decodificar y verificar tokens JWT

/**
 * Decodifica un token JWT sin verificar la firma
 * @param {string} token - Token JWT
 * @returns {object|null} - Payload del token o null si es inválido
 */
export const decodeToken = (token) => {
  try {
    if (!token) return null;
    
    const parts = token.split('.');
    if (parts.length !== 3) return null;
    
    const payload = parts[1];
    const decoded = JSON.parse(atob(payload.replace(/-/g, '+').replace(/_/g, '/')));
    return decoded;
  } catch (error) {
    console.error('Error decodificando token:', error);
    return null;
  }
};

/**
 * Verifica si un token JWT está expirado
 * @param {string} token - Token JWT
 * @returns {boolean} - true si está expirado, false si no
 */
export const isTokenExpired = (token) => {
  const decoded = decodeToken(token);
  if (!decoded || !decoded.exp) return true;
  
  const currentTime = Math.floor(Date.now() / 1000);
  return decoded.exp < currentTime;
};

/**
 * Obtiene el tiempo restante hasta la expiración del token en segundos
 * @param {string} token - Token JWT
 * @returns {number} - Segundos restantes hasta la expiración, o 0 si está expirado
 */
export const getTokenTimeRemaining = (token) => {
  const decoded = decodeToken(token);
  if (!decoded || !decoded.exp) return 0;
  
  const currentTime = Math.floor(Date.now() / 1000);
  const remaining = decoded.exp - currentTime;
  return remaining > 0 ? remaining : 0;
};

