import { useLocation, useNavigate } from 'react-router-dom';
import { useEffect } from 'react';

const Congrats = () => {
  const location = useLocation();
  const navigate = useNavigate();
  const { reserva, cancha } = location.state || {};

  useEffect(() => {
    if (!reserva) {
      navigate('/');
    }
  }, [reserva, navigate]);

  if (!reserva) {
    return null;
  }

  return (
    <div style={styles.container}>
      <div style={styles.card}>
        <div style={styles.iconContainer}>
          <span style={styles.icon}>✓</span>
        </div>

        <h1 style={styles.title}>¡Reserva Confirmada!</h1>
        <p style={styles.subtitle}>
          Tu reserva ha sido creada exitosamente
        </p>

        <div style={styles.details}>
          <h3 style={styles.detailsTitle}>Detalles de la Reserva</h3>

          <div style={styles.detailRow}>
            <span style={styles.label}>Cancha:</span>
            <span style={styles.value}>{reserva.cancha_name || cancha?.name}</span>
          </div>

          <div style={styles.detailRow}>
            <span style={styles.label}>Fecha:</span>
            <span style={styles.value}>{reserva.date}</span>
          </div>

          <div style={styles.detailRow}>
            <span style={styles.label}>Horario:</span>
            <span style={styles.value}>
              {reserva.start_time} - {reserva.end_time}
            </span>
          </div>

          <div style={styles.detailRow}>
            <span style={styles.label}>Duración:</span>
            <span style={styles.value}>{reserva.duration} minutos</span>
          </div>

          <div style={styles.detailRow}>
            <span style={styles.label}>Estado:</span>
            <span style={{...styles.value, ...styles.statusBadge}}>
              {reserva.status}
            </span>
          </div>

          <div style={{...styles.detailRow, ...styles.totalRow}}>
            <span style={styles.label}>Total:</span>
            <span style={styles.price}>${reserva.total_price}</span>
          </div>
        </div>

        <div style={styles.info}>
          <p style={styles.infoText}>
            📧 Hemos enviado un correo de confirmación con todos los detalles
          </p>
          <p style={styles.infoText}>
            💡 Recuerda llegar 10 minutos antes de tu horario reservado
          </p>
        </div>

        <div style={styles.actions}>
          <button onClick={() => navigate('/')} style={styles.homeBtn}>
            Volver al Inicio
          </button>
          <button onClick={() => navigate('/mis-reservas')} style={styles.reservasBtn}>
            Ver Mis Reservas
          </button>
        </div>
      </div>
    </div>
  );
};

const styles = {
  container: {
    minHeight: '100vh',
    backgroundColor: '#f5f5f5',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    padding: '1rem',
  },
  card: {
    backgroundColor: '#fff',
    padding: '3rem',
    borderRadius: '12px',
    boxShadow: '0 4px 12px rgba(0,0,0,0.1)',
    maxWidth: '600px',
    width: '100%',
    textAlign: 'center',
  },
  iconContainer: {
    marginBottom: '1.5rem',
  },
  icon: {
    display: 'inline-block',
    width: '80px',
    height: '80px',
    borderRadius: '50%',
    backgroundColor: '#27ae60',
    color: '#fff',
    fontSize: '3rem',
    lineHeight: '80px',
    fontWeight: 'bold',
  },
  title: {
    fontSize: '2rem',
    marginBottom: '0.5rem',
    color: '#2c3e50',
  },
  subtitle: {
    fontSize: '1.1rem',
    color: '#7f8c8d',
    marginBottom: '2rem',
  },
  details: {
    backgroundColor: '#f8f9fa',
    padding: '1.5rem',
    borderRadius: '8px',
    marginBottom: '1.5rem',
    textAlign: 'left',
  },
  detailsTitle: {
    fontSize: '1.3rem',
    marginBottom: '1rem',
    color: '#2c3e50',
    textAlign: 'center',
  },
  detailRow: {
    display: 'flex',
    justifyContent: 'space-between',
    padding: '0.75rem 0',
    borderBottom: '1px solid #ecf0f1',
  },
  label: {
    fontWeight: '500',
    color: '#7f8c8d',
  },
  value: {
    color: '#2c3e50',
    fontWeight: '500',
  },
  statusBadge: {
    backgroundColor: '#27ae60',
    color: '#fff',
    padding: '0.25rem 0.75rem',
    borderRadius: '12px',
    fontSize: '0.9rem',
    textTransform: 'capitalize',
  },
  totalRow: {
    borderBottom: 'none',
    paddingTop: '1rem',
    marginTop: '0.5rem',
    borderTop: '2px solid #2c3e50',
  },
  price: {
    fontSize: '1.5rem',
    fontWeight: 'bold',
    color: '#27ae60',
  },
  info: {
    marginBottom: '2rem',
  },
  infoText: {
    fontSize: '0.95rem',
    color: '#7f8c8d',
    marginBottom: '0.5rem',
  },
  actions: {
    display: 'flex',
    gap: '1rem',
    justifyContent: 'center',
  },
  homeBtn: {
    backgroundColor: '#95a5a6',
    color: '#fff',
    padding: '0.75rem 1.5rem',
    border: 'none',
    borderRadius: '4px',
    cursor: 'pointer',
    fontSize: '1rem',
    fontWeight: '500',
  },
  reservasBtn: {
    backgroundColor: '#3498db',
    color: '#fff',
    padding: '0.75rem 1.5rem',
    border: 'none',
    borderRadius: '4px',
    cursor: 'pointer',
    fontSize: '1rem',
    fontWeight: '500',
  },
};

export default Congrats;