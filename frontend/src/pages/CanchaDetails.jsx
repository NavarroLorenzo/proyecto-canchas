import { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import canchaService from '../services/canchaService';
import reservaService from '../services/reservaService';

const CanchaDetails = () => {
  const { id } = useParams();
  const navigate = useNavigate();
  const { user, token, isAuthenticated } = useAuth();

  const [cancha, setCancha] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  // Formulario de reserva
  const [reservaData, setReservaData] = useState({
    date: '',
    start_time: '',
    end_time: '',
  });
  const [reservaLoading, setReservaLoading] = useState(false);
  const [reservaError, setReservaError] = useState('');

  useEffect(() => {
    fetchCancha();
  }, [id]);

  const fetchCancha = async () => {
    try {
      const data = await canchaService.getCanchaById(id);
      setCancha(data);
    } catch (err) {
      setError('Error al cargar los detalles de la cancha');
    } finally {
      setLoading(false);
    }
  };

  const handleReservaChange = (e) => {
    setReservaData({
      ...reservaData,
      [e.target.name]: e.target.value,
    });
  };

  const handleReservar = async (e) => {
    e.preventDefault();

    if (!isAuthenticated) {
      navigate('/login', { state: { from: `/cancha/${id}` } });
      return;
    }

    setReservaError('');
    setReservaLoading(true);

    try {
      const payload = {
        cancha_id: id,
        user_id: user.id,
        date: reservaData.date,
        start_time: reservaData.start_time,
        end_time: reservaData.end_time,
      };

      const response = await reservaService.createReserva(payload, token);
      
      // Redirigir a página de confirmación
      navigate('/congrats', { 
        state: { 
          reserva: response,
          cancha: cancha 
        } 
      });
    } catch (err) {
      setReservaError(
        err.response?.data?.message || 'Error al crear la reserva'
      );
    } finally {
      setReservaLoading(false);
    }
  };

  const calculateDuration = () => {
    if (reservaData.start_time && reservaData.end_time) {
      const start = new Date(`2000-01-01T${reservaData.start_time}`);
      const end = new Date(`2000-01-01T${reservaData.end_time}`);
      const diff = (end - start) / (1000 * 60); // minutos
      return diff > 0 ? diff : 0;
    }
    return 0;
  };

  const calculatePrice = () => {
    const duration = calculateDuration();
    if (duration > 0 && cancha) {
      const hours = duration / 60;
      return (cancha.price * hours).toFixed(2);
    }
    return 0;
  };

  if (loading) {
    return <div style={styles.loading}>Cargando...</div>;
  }

  if (error) {
    return (
      <div style={styles.container}>
        <div style={styles.error}>{error}</div>
        <button onClick={() => navigate('/')} style={styles.backBtn}>
          Volver al inicio
        </button>
      </div>
    );
  }

  if (!cancha) {
    return (
      <div style={styles.container}>
        <div style={styles.error}>Cancha no encontrada</div>
        <button onClick={() => navigate('/')} style={styles.backBtn}>
          Volver al inicio
        </button>
      </div>
    );
  }

  return (
    <div style={styles.container}>
      <button onClick={() => navigate('/')} style={styles.backBtn}>
        ← Volver
      </button>

      <div style={styles.content}>
        {/* Detalles de la cancha */}
        <div style={styles.leftSection}>
          <div style={styles.imageContainer}>
            {cancha.image_url ? (
              <img
                src={cancha.image_url}
                alt={cancha.name}
                style={styles.image}
              />
            ) : (
              <div style={styles.imagePlaceholder}>
                {getTypeEmoji(cancha.type)}
              </div>
            )}
          </div>

          <h1 style={styles.title}>{cancha.name}</h1>

          <div style={styles.badges}>
            <span style={styles.badge}>{cancha.type}</span>
            <span style={styles.badge}>👥 {cancha.capacity} personas</span>
            <span style={{...styles.badge, backgroundColor: cancha.available ? '#27ae60' : '#e74c3c'}}>
              {cancha.available ? '✓ Disponible' : '✗ No disponible'}
            </span>
          </div>

          <div style={styles.infoSection}>
            <h3 style={styles.sectionTitle}>📍 Ubicación</h3>
            <p style={styles.text}>{cancha.location}</p>
            <p style={styles.text}>{cancha.address}</p>
          </div>

          <div style={styles.infoSection}>
            <h3 style={styles.sectionTitle}>📝 Descripción</h3>
            <p style={styles.text}>{cancha.description}</p>
          </div>

          <div style={styles.priceSection}>
            <span style={styles.priceLabel}>Precio por hora:</span>
            <span style={styles.price}>${cancha.price}</span>
          </div>
        </div>

        {/* Formulario de reserva */}
        <div style={styles.rightSection}>
          <div style={styles.reservaCard}>
            <h2 style={styles.reservaTitle}>Reservar Cancha</h2>

            {!isAuthenticated && (
              <div style={styles.warning}>
                Debes <a href="/login" style={styles.link}>iniciar sesión</a> para reservar
              </div>
            )}

            {reservaError && (
              <div style={styles.errorBox}>{reservaError}</div>
            )}

            <form onSubmit={handleReservar} style={styles.form}>
              <div style={styles.formGroup}>
                <label style={styles.label}>Fecha</label>
                <input
                  type="date"
                  name="date"
                  value={reservaData.date}
                  onChange={handleReservaChange}
                  required
                  min={new Date().toISOString().split('T')[0]}
                  style={styles.input}
                  disabled={!cancha.available}
                />
              </div>

              <div style={styles.formGroup}>
                <label style={styles.label}>Hora de inicio</label>
                <input
                  type="time"
                  name="start_time"
                  value={reservaData.start_time}
                  onChange={handleReservaChange}
                  required
                  style={styles.input}
                  disabled={!cancha.available}
                />
              </div>

              <div style={styles.formGroup}>
                <label style={styles.label}>Hora de fin</label>
                <input
                  type="time"
                  name="end_time"
                  value={reservaData.end_time}
                  onChange={handleReservaChange}
                  required
                  style={styles.input}
                  disabled={!cancha.available}
                />
              </div>

              {calculateDuration() > 0 && (
                <div style={styles.summary}>
                  <div style={styles.summaryRow}>
                    <span>Duración:</span>
                    <span>{calculateDuration()} minutos</span>
                  </div>
                  <div style={styles.summaryRow}>
                    <span>Precio por hora:</span>
                    <span>${cancha.price}</span>
                  </div>
                  <div style={{...styles.summaryRow, ...styles.summaryTotal}}>
                    <span>Total:</span>
                    <span>${calculatePrice()}</span>
                  </div>
                </div>
              )}

              <button
                type="submit"
                disabled={reservaLoading || !cancha.available || !isAuthenticated}
                style={{
                  ...styles.submitBtn,
                  ...(reservaLoading || !cancha.available || !isAuthenticated ? styles.submitBtnDisabled : {})
                }}
              >
                {reservaLoading ? 'Reservando...' : 'Confirmar Reserva'}
              </button>
            </form>
          </div>
        </div>
      </div>
    </div>
  );
};

const getTypeEmoji = (type) => {
  const emojis = {
    futbol: '⚽',
    tenis: '🎾',
    basquet: '🏀',
    paddle: '🏓',
    voley: '🏐',
  };
  return emojis[type] || '🏟️';
};

const styles = {
  container: {
    maxWidth: '1200px',
    margin: '0 auto',
    padding: '2rem 1rem',
  },
  loading: {
    textAlign: 'center',
    padding: '3rem',
    fontSize: '1.2rem',
  },
  error: {
    backgroundColor: '#fee',
    color: '#c00',
    padding: '1rem',
    borderRadius: '4px',
    marginBottom: '1rem',
  },
  backBtn: {
    backgroundColor: '#95a5a6',
    color: '#fff',
    padding: '0.5rem 1rem',
    border: 'none',
    borderRadius: '4px',
    cursor: 'pointer',
    marginBottom: '1.5rem',
  },
  content: {
    display: 'grid',
    gridTemplateColumns: '1fr 400px',
    gap: '2rem',
  },
  leftSection: {
    backgroundColor: '#fff',
    padding: '2rem',
    borderRadius: '8px',
    boxShadow: '0 2px 8px rgba(0,0,0,0.1)',
  },
  imageContainer: {
    width: '100%',
    height: '400px',
    marginBottom: '1.5rem',
    borderRadius: '8px',
    overflow: 'hidden',
  },
  image: {
    width: '100%',
    height: '100%',
    objectFit: 'cover',
  },
  imagePlaceholder: {
    width: '100%',
    height: '100%',
    backgroundColor: '#ecf0f1',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    fontSize: '6rem',
  },
  title: {
    fontSize: '2rem',
    marginBottom: '1rem',
    color: '#2c3e50',
  },
  badges: {
    display: 'flex',
    gap: '0.5rem',
    marginBottom: '1.5rem',
    flexWrap: 'wrap',
  },
  badge: {
    backgroundColor: '#3498db',
    color: '#fff',
    padding: '0.5rem 1rem',
    borderRadius: '20px',
    fontSize: '0.9rem',
    textTransform: 'capitalize',
  },
  infoSection: {
    marginBottom: '1.5rem',
  },
  sectionTitle: {
    fontSize: '1.3rem',
    marginBottom: '0.5rem',
    color: '#2c3e50',
  },
  text: {
    fontSize: '1rem',
    color: '#7f8c8d',
    lineHeight: '1.6',
    marginBottom: '0.5rem',
  },
  priceSection: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
    padding: '1.5rem',
    backgroundColor: '#ecf0f1',
    borderRadius: '8px',
    marginTop: '2rem',
  },
  priceLabel: {
    fontSize: '1.2rem',
    color: '#2c3e50',
  },
  price: {
    fontSize: '2rem',
    fontWeight: 'bold',
    color: '#27ae60',
  },
  rightSection: {
    position: 'sticky',
    top: '1rem',
    height: 'fit-content',
  },
  reservaCard: {
    backgroundColor: '#fff',
    padding: '2rem',
    borderRadius: '8px',
    boxShadow: '0 2px 8px rgba(0,0,0,0.1)',
  },
  reservaTitle: {
    fontSize: '1.5rem',
    marginBottom: '1.5rem',
    color: '#2c3e50',
  },
  warning: {
    backgroundColor: '#fff3cd',
    color: '#856404',
    padding: '0.75rem',
    borderRadius: '4px',
    marginBottom: '1rem',
    fontSize: '0.9rem',
  },
  errorBox: {
    backgroundColor: '#fee',
    color: '#c00',
    padding: '0.75rem',
    borderRadius: '4px',
    marginBottom: '1rem',
    fontSize: '0.9rem',
  },
  form: {
    display: 'flex',
    flexDirection: 'column',
    gap: '1rem',
  },
  formGroup: {
    display: 'flex',
    flexDirection: 'column',
    gap: '0.5rem',
  },
  label: {
    fontWeight: '500',
    color: '#2c3e50',
  },
  input: {
    padding: '0.75rem',
    border: '1px solid #ddd',
    borderRadius: '4px',
    fontSize: '1rem',
  },
  summary: {
    backgroundColor: '#f8f9fa',
    padding: '1rem',
    borderRadius: '4px',
    marginTop: '0.5rem',
  },
  summaryRow: {
    display: 'flex',
    justifyContent: 'space-between',
    marginBottom: '0.5rem',
    fontSize: '0.95rem',
    color: '#2c3e50',
  },
  summaryTotal: {
    borderTop: '1px solid #ddd',
    paddingTop: '0.5rem',
    marginTop: '0.5rem',
    fontWeight: 'bold',
    fontSize: '1.1rem',
  },
  submitBtn: {
    backgroundColor: '#27ae60',
    color: '#fff',
    padding: '0.75rem',
    border: 'none',
    borderRadius: '4px',
    fontSize: '1rem',
    fontWeight: '500',
    cursor: 'pointer',
    marginTop: '0.5rem',
  },
  submitBtnDisabled: {
    backgroundColor: '#bdc3c7',
    cursor: 'not-allowed',
  },
  link: {
    color: '#3498db',
    textDecoration: 'underline',
  },
};

export default CanchaDetails;