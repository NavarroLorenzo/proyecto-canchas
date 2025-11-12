import { useState, useEffect } from 'react';
import { useAuth } from '../context/AuthContext';
import { useNavigate } from 'react-router-dom';
import reservaService from '../services/reservaService';

const MisReservas = () => {
  const { user, token } = useAuth();
  const navigate = useNavigate();

  const [reservas, setReservas] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    fetchReservas();
  }, []);

  const fetchReservas = async () => {
    try {
      const response = await reservaService.getReservasByUserId(user.id, token);
      setReservas(response.reservas || []);
    } catch (err) {
      setError('Error al cargar las reservas');
    } finally {
      setLoading(false);
    }
  };

  const handleCancelReserva = async (id) => {
    if (!window.confirm('¿Estás seguro de que deseas cancelar esta reserva?')) {
      return;
    }

    try {
      await reservaService.cancelReserva(id, token);
      // Actualizar la lista
      fetchReservas();
    } catch (err) {
      alert('Error al cancelar la reserva');
    }
  };

  const getStatusColor = (status) => {
    const colors = {
      confirmed: '#27ae60',
      pending: '#f39c12',
      cancelled: '#e74c3c',
    };
    return colors[status] || '#95a5a6';
  };

  const getStatusText = (status) => {
    const texts = {
      confirmed: 'Confirmada',
      pending: 'Pendiente',
      cancelled: 'Cancelada',
    };
    return texts[status] || status;
  };

  if (loading) {
    return <div style={styles.loading}>Cargando reservas...</div>;
  }

  return (
    <div style={styles.container}>
      <h1 style={styles.title}>Mis Reservas</h1>

      {error && <div style={styles.error}>{error}</div>}

      {reservas.length === 0 ? (
        <div style={styles.empty}>
          <p style={styles.emptyText}>No tienes reservas activas</p>
          <button onClick={() => navigate('/')} style={styles.searchBtn}>
            Buscar Canchas
          </button>
        </div>
      ) : (
        <div style={styles.grid}>
          {reservas.map((reserva) => (
            <div key={reserva.id} style={styles.card}>
              <div style={styles.cardHeader}>
                <h3 style={styles.cardTitle}>{reserva.cancha_name}</h3>
                <span
                  style={{
                    ...styles.statusBadge,
                    backgroundColor: getStatusColor(reserva.status),
                  }}
                >
                  {getStatusText(reserva.status)}
                </span>
              </div>

              <div style={styles.cardBody}>
                <div style={styles.infoRow}>
                  <span style={styles.infoLabel}>📅 Fecha:</span>
                  <span style={styles.infoValue}>{reserva.date}</span>
                </div>

                <div style={styles.infoRow}>
                  <span style={styles.infoLabel}>🕐 Horario:</span>
                  <span style={styles.infoValue}>
                    {reserva.start_time} - {reserva.end_time}
                  </span>
                </div>

                <div style={styles.infoRow}>
                  <span style={styles.infoLabel}>⏱️ Duración:</span>
                  <span style={styles.infoValue}>{reserva.duration} min</span>
                </div>

                <div style={styles.infoRow}>
                  <span style={styles.infoLabel}>💰 Total:</span>
                  <span style={styles.price}>${reserva.total_price}</span>
                </div>
              </div>

              <div style={styles.cardFooter}>
                <button
                  onClick={() => navigate(`/cancha/${reserva.cancha_id}`)}
                  style={styles.detailsBtn}
                >
                  Ver Cancha
                </button>
                {reserva.status !== 'cancelled' && (
                  <button
                    onClick={() => handleCancelReserva(reserva.id)}
                    style={styles.cancelBtn}
                  >
                    Cancelar
                  </button>
                )}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};

const styles = {
  container: {
    maxWidth: '1200px',
    margin: '0 auto',
    padding: '2rem 1rem',
  },
  title: {
    fontSize: '2rem',
    marginBottom: '2rem',
    color: '#2c3e50',
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
  empty: {
    textAlign: 'center',
    padding: '3rem',
  },
  emptyText: {
    fontSize: '1.2rem',
    color: '#7f8c8d',
    marginBottom: '1.5rem',
  },
  searchBtn: {
    backgroundColor: '#3498db',
    color: '#fff',
    padding: '0.75rem 1.5rem',
    border: 'none',
    borderRadius: '4px',
    cursor: 'pointer',
    fontSize: '1rem',
  },
  grid: {
    display: 'grid',
    gridTemplateColumns: 'repeat(auto-fill, minmax(320px, 1fr))',
    gap: '1.5rem',
  },
  card: {
    backgroundColor: '#fff',
    borderRadius: '8px',
    boxShadow: '0 2px 8px rgba(0,0,0,0.1)',
    overflow: 'hidden',
  },
  cardHeader: {
    padding: '1.5rem',
    borderBottom: '1px solid #ecf0f1',
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
  },
  cardTitle: {
    fontSize: '1.3rem',
    color: '#2c3e50',
    margin: 0,
  },
  statusBadge: {
    color: '#fff',
    padding: '0.25rem 0.75rem',
    borderRadius: '12px',
    fontSize: '0.85rem',
    fontWeight: '500',
  },
  cardBody: {
    padding: '1.5rem',
  },
  infoRow: {
    display: 'flex',
    justifyContent: 'space-between',
    marginBottom: '0.75rem',
  },
  infoLabel: {
    color: '#7f8c8d',
  },
  infoValue: {
    color: '#2c3e50',
    fontWeight: '500',
  },
  price: {
    fontSize: '1.3rem',
    fontWeight: 'bold',
    color: '#27ae60',
  },
  cardFooter: {
    padding: '1rem 1.5rem',
    backgroundColor: '#f8f9fa',
    display: 'flex',
    gap: '0.75rem',
  },
  detailsBtn: {
    flex: 1,
    backgroundColor: '#3498db',
    color: '#fff',
    padding: '0.5rem',
    border: 'none',
    borderRadius: '4px',
    cursor: 'pointer',
  },
  cancelBtn: {
    flex: 1,
    backgroundColor: '#e74c3c',
    color: '#fff',
    padding: '0.5rem',
    border: 'none',
    borderRadius: '4px',
    cursor: 'pointer',
  },
};

export default MisReservas;