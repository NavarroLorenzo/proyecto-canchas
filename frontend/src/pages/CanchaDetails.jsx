import { useState, useEffect, useMemo } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import canchaService from '../services/canchaService';
import reservaService from '../services/reservaService';

const PADDEL_TYPES = ['padel', 'tenis', 'paddle'];
const START_MINUTES = 10 * 60;
const END_MINUTES = 26 * 60;
const MINUTES_IN_DAY = 24 * 60;

const isExtendedSlot = (type = '') =>
  PADDEL_TYPES.includes(type.toLowerCase());

const minutesToLabel = (minutes) => {
  const normalized = ((minutes % MINUTES_IN_DAY) + MINUTES_IN_DAY) % MINUTES_IN_DAY;
  const hours = Math.floor(normalized / 60);
  const mins = normalized % 60;
  return `${hours.toString().padStart(2, '0')}:${mins.toString().padStart(2, '0')}`;
};

const parseTimeToMinutes = (timeStr) => {
  if (!timeStr) return 0;
  const [hours, minutes] = timeStr.split(':').map(Number);
  return hours * 60 + minutes;
};

const normalizeMinutesForRange = (timeStr) => {
  let minutes = parseTimeToMinutes(timeStr);
  if (minutes < START_MINUTES) {
    minutes += MINUTES_IN_DAY;
  }
  return minutes;
};

const generateSlots = (type = '') => {
  const duration = isExtendedSlot(type) ? 90 : 60;
  const slots = [];
  for (let current = START_MINUTES; current + duration <= END_MINUTES; current += duration) {
    const start = minutesToLabel(current);
    const end = minutesToLabel(current + duration);
    slots.push({
      key: `${start}-${end}`,
      start,
      end,
      duration,
      normalizedStart: current,
      normalizedEnd: current + duration,
    });
  }
  return slots;
};

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
  const [bookedSlots, setBookedSlots] = useState([]);
  const [slotsLoading, setSlotsLoading] = useState(false);
  const [slotsError, setSlotsError] = useState('');

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

  const slotOptions = useMemo(() => {
    if (!cancha) return [];
    return generateSlots(cancha.type || '');
  }, [cancha]);

  const slotsWithStatus = useMemo(
    () =>
      slotOptions.map((slot) => {
        const isBooked = bookedSlots.some(
          (booked) => booked.normalizedStart === slot.normalizedStart
        );
        const isSelected =
          reservaData.start_time === slot.start &&
          reservaData.end_time === slot.end;
        return { ...slot, isBooked, isSelected };
      }),
    [slotOptions, bookedSlots, reservaData.start_time, reservaData.end_time]
  );

  useEffect(() => {
    if (!reservaData.date || !id) {
      setBookedSlots([]);
      setSlotsError('');
      return;
    }

    const loadBookedSlots = async () => {
      setSlotsLoading(true);
      setSlotsError('');
      try {
        const data = await reservaService.getReservasByCanchaId(id);
        const reservas = data?.reservas || [];
        const sameDay = reservas.filter(
          (item) => item.date === reservaData.date && item.status !== 'cancelled'
        );
        const normalized = sameDay.map((item) => {
          const normalizedStart = normalizeMinutesForRange(item.start_time);
          const normalizedEnd = normalizeMinutesForRange(item.end_time);
          return {
            start: item.start_time,
            end: item.end_time,
            normalizedStart,
            normalizedEnd:
              normalizedEnd <= normalizedStart
                ? normalizedEnd + MINUTES_IN_DAY
                : normalizedEnd,
          };
        });
        setBookedSlots(normalized);
      } catch (err) {
        console.error('Error fetching booked slots:', err);
        setBookedSlots([]);
        setSlotsError('No se pudieron cargar los turnos. Intenta nuevamente.');
      } finally {
        setSlotsLoading(false);
      }
    };

    loadBookedSlots();
  }, [id, reservaData.date]);

  const handleReservaChange = (e) => {
    const { name, value } = e.target;
    setReservaData((prev) => ({
      ...prev,
      [name]: value,
      ...(name === 'date'
        ? {
            start_time: '',
            end_time: '',
          }
        : {}),
    }));
  };

  const handleSelectSlot = (slot) => {
    if (slot.isBooked) return;
    setReservaError('');
    setReservaData((prev) => ({
      ...prev,
      start_time: slot.start,
      end_time: slot.end,
    }));
  };

  const handleReservar = async (e) => {
    e.preventDefault();

    if (!isAuthenticated) {
      navigate('/login', { state: { from: `/cancha/${id}` } });
      return;
    }

    if (!reservaData.date || !reservaData.start_time || !reservaData.end_time) {
      setReservaError('Selecciona una fecha y un horario disponible.');
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
      const start = normalizeMinutesForRange(reservaData.start_time);
      let end = normalizeMinutesForRange(reservaData.end_time);
      if (end <= start) {
        end += MINUTES_IN_DAY;
      }
      return end - start;
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

  const isFormReady = Boolean(
    reservaData.date && reservaData.start_time && reservaData.end_time
  );

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
            <span style={{ ...styles.badge, backgroundColor: cancha.available ? '#27ae60' : '#e74c3c' }}>
              {cancha.available ? '✓ Disponible' : '✗ No disponible'}
            </span>
          </div>

          <div style={styles.infoSection}>
            <h3 style={styles.sectionTitle}>N° Cancha</h3>
            <p style={styles.text}>#{cancha.number || cancha.id}</p>
          </div>

          <div style={styles.infoSection}>
            <h3 style={styles.sectionTitle}>📝 Descripción</h3>
            <p style={styles.text}>{cancha.description}</p>
          </div>

          <div style={styles.priceSection}>
            <span style={styles.priceLabel}>Precio por turno:</span>
            <span style={styles.price}>${cancha.price}</span>
          </div>
          <p style={styles.priceNote}>
            El precio mostrado incluye 5% de impuestos y un fee de mantenimiento.
          </p>
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
                <label style={styles.label}>Turnos disponibles</label>
                {!reservaData.date ? (
                  <div style={styles.slotInfo}>
                    Selecciona una fecha para ver los horarios disponibles.
                  </div>
                ) : slotsLoading ? (
                  <div style={styles.slotInfo}>Cargando turnos...</div>
                ) : (
                  <>
                    {slotsError && (
                      <div style={styles.slotError}>{slotsError}</div>
                    )}

                    <div style={styles.slotLegend}>
                      <span style={styles.legendItem}>
                        <span
                          style={{
                            ...styles.legendDot,
                            backgroundColor: '#27ae60',
                          }}
                        ></span>
                        Disponible
                      </span>
                      <span style={styles.legendItem}>
                        <span
                          style={{
                            ...styles.legendDot,
                            backgroundColor: '#bdc3c7',
                          }}
                        ></span>
                        Ocupado
                      </span>
                      <span style={styles.legendItem}>
                        <span
                          style={{
                            ...styles.legendDot,
                            backgroundColor: '#2980b9',
                          }}
                        ></span>
                        Seleccionado
                      </span>
                    </div>

                    {slotsWithStatus.length === 0 ? (
                      <div style={styles.slotInfo}>
                        No hay turnos configurados para esta cancha.
                      </div>
                    ) : (
                      <div style={styles.slotGrid}>
                        {slotsWithStatus.map((slot) => (
                          <button
                            key={slot.key}
                            type="button"
                            onClick={() => handleSelectSlot(slot)}
                            disabled={slot.isBooked || !cancha.available}
                            style={{
                              ...styles.slotButton,
                              ...(slot.isBooked ? styles.slotButtonBooked : {}),
                              ...(slot.isSelected ? styles.slotButtonSelected : {}),
                            }}
                          >
                            {slot.start} - {slot.end}
                          </button>
                        ))}
                      </div>
                    )}
                  </>
                )}
              </div>

              {isFormReady && (
                <div style={styles.summary}>
                  <div style={styles.summaryRow}>
                    <span>Horario seleccionado:</span>
                    <span>
                      {reservaData.start_time} - {reservaData.end_time}
                    </span>
                  </div>
                  <div style={styles.summaryRow}>
                    <span>Duración:</span>
                    <span>{calculateDuration()} minutos</span>
                  </div>
                  <div style={styles.summaryRow}>
                    <span>Precio por turno:</span>
                    <span>${cancha.price}</span>
                  </div>
                  <div style={{ ...styles.summaryRow, ...styles.summaryTotal }}>
                    <span>Total:</span>
                    <span>${calculatePrice()}</span>
                  </div>
                </div>
              )}

              <button
                type="submit"
                disabled={
                  reservaLoading ||
                  !cancha.available ||
                  !isAuthenticated ||
                  !isFormReady
                }
                style={{
                  ...styles.submitBtn,
                  ...(reservaLoading ||
                  !cancha.available ||
                  !isAuthenticated ||
                  !isFormReady
                    ? styles.submitBtnDisabled
                    : {}),
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
  priceNote: {
    marginTop: '0.5rem',
    fontSize: '0.95rem',
    color: '#7f8c8d',
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
  slotLegend: {
    display: 'flex',
    flexWrap: 'wrap',
    gap: '1rem',
    fontSize: '0.85rem',
    color: '#7f8c8d',
    marginBottom: '0.75rem',
  },
  legendItem: {
    display: 'flex',
    alignItems: 'center',
    gap: '0.4rem',
  },
  legendDot: {
    width: '12px',
    height: '12px',
    borderRadius: '50%',
    display: 'inline-block',
  },
  slotGrid: {
    display: 'grid',
    gridTemplateColumns: 'repeat(auto-fill, minmax(120px, 1fr))',
    gap: '0.5rem',
  },
  slotButton: {
    border: '1px solid #dfe6e9',
    borderRadius: '6px',
    padding: '0.5rem 0.25rem',
    backgroundColor: '#fff',
    cursor: 'pointer',
    fontWeight: '500',
    color: '#2c3e50',
    transition: 'all 0.2s',
  },
  slotButtonSelected: {
    backgroundColor: '#2980b9',
    color: '#fff',
    borderColor: '#2980b9',
  },
  slotButtonBooked: {
    backgroundColor: '#ecf0f1',
    color: '#95a5a6',
    borderColor: '#ecf0f1',
    cursor: 'not-allowed',
  },
  slotInfo: {
    backgroundColor: '#f8f9fa',
    padding: '0.75rem',
    borderRadius: '4px',
    fontSize: '0.9rem',
    color: '#7f8c8d',
  },
  slotError: {
    backgroundColor: '#fee',
    color: '#c0392b',
    padding: '0.5rem',
    borderRadius: '4px',
    marginBottom: '0.5rem',
    fontSize: '0.9rem',
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
