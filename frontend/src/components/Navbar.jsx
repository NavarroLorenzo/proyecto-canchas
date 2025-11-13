import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';

const Navbar = () => {
  const { user, logout, isAuthenticated, isAdmin } = useAuth();
  const navigate = useNavigate();

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  return (
    <nav style={styles.nav}>
      <div style={styles.container}>
        <Link to="/" style={styles.logo}>
          Canchas Deportivas
        </Link>

        <div style={styles.menu}>
          {isAuthenticated ? (
            <>
              <Link to="/" style={styles.link}>
                Inicio
              </Link>
              <Link to="/mis-reservas" style={styles.link}>
                Mis Reservas
              </Link>
              {isAdmin() && (
                <Link to="/admin" style={styles.link}>
                  Administración
                </Link>
              )}
              <span style={styles.user}>👤 {user?.username}</span>
              <button onClick={handleLogout} style={styles.logoutBtn}>
                Cerrar Sesión
              </button>
            </>
          ) : (
            <>
              <Link to="/login" style={styles.link}>
                Iniciar Sesión
              </Link>
              <Link to="/register" style={styles.registerBtn}>
                Registrarse
              </Link>
            </>
          )}
        </div>
      </div>
    </nav>
  );
};

const styles = {
  nav: {
    backgroundColor: '#2c3e50',
    padding: '1rem 0',
    boxShadow: '0 2px 4px rgba(0,0,0,0.1)',
  },
  container: {
    maxWidth: '1200px',
    margin: '0 auto',
    padding: '0 1rem',
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
  },
  logo: {
    color: '#fff',
    fontSize: '1.5rem',
    fontWeight: 'bold',
    textDecoration: 'none',
  },
  menu: {
    display: 'flex',
    gap: '1.5rem',
    alignItems: 'center',
  },
  link: {
    color: '#ecf0f1',
    textDecoration: 'none',
    fontSize: '1rem',
    transition: 'color 0.3s',
  },
  user: {
    color: '#ecf0f1',
    fontSize: '0.9rem',
  },
  logoutBtn: {
    backgroundColor: '#e74c3c',
    color: '#fff',
    border: 'none',
    padding: '0.5rem 1rem',
    borderRadius: '4px',
    cursor: 'pointer',
    fontSize: '0.9rem',
  },
  registerBtn: {
    backgroundColor: '#27ae60',
    color: '#fff',
    padding: '0.5rem 1rem',
    borderRadius: '4px',
    textDecoration: 'none',
    fontSize: '0.9rem',
  },
};

export default Navbar;