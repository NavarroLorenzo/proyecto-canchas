import { useState, useEffect } from 'react';
import { useAuth } from '../context/AuthContext';
import canchaService from '../services/canchaService';
import reservaService from '../services/reservaService';

const Admin = () => {
  const { token } = useAuth();
  const [activeTab, setActiveTab] = useState('canchas'); // 'canchas' o 'reservas'

  // Estado para canchas
  const [canchas, setCanchas] = useState([]);
  const [canchasLoading, setCanchasLoading] = useState(true);
  
  // Estado para reservas
  const [reservas, setReservas] = useState([]);
  const [reservasLoading, setReservasLoading] = useState(true);

  // Modal para crear/editar cancha
  const [showCanchaModal, setShowCanchaModal] = useState(false);
  const [editingCancha, setEditingCancha] = useState(null);
  const [canchaForm, setCanchaForm] = useState({
    name: '',
    type: 'futbol',
    description: '',
    location: '',
    address: '',
    price: '',
    capacity: '',
    available: true,
    image_url: '',
  });

  useEffect(() => {
    fetchCanchas();
    fetchReservas();
  }, []);

  const fetchCanchas = async () => {
    try {
      const response = await canchaService.getAllCanchas();
      setCanchas(response.canchas || []);
    } catch (err) {
      console.error('Error fetching canchas:', err);
    } finally {
      setCanchasLoading(false);
    }
  };

  const fetchReservas = async () => {
    try {
      const response = await reservaService.getAllReservas(token);
      setReservas(response.reservas || []);
    } catch (err) {
      console.error('Error fetching reservas:', err);
    } finally {
      setReservasLoading(false);
    }
  };

  // CRUD de Canchas
  const handleCreateCancha = () => {
    setEditingCancha(null);
    setCanchaForm({
      name: '',
      type: 'futbol',
      description: '',
      location: '',
      address: '',
      price: '',
      capacity: '',
      available: true,
      image_url: '',
    });
    setShowCanchaModal(true);
  };

  const handleEditCancha = (cancha) => {
    setEditingCancha(cancha);
    setCanchaForm({
      name: cancha.name,
      type: cancha.type,
      description: cancha.description,
      location: cancha.location,
      address: cancha.address,
      price: cancha.price.toString(),
      capacity: cancha.capacity.toString(),
      available: cancha.available,
      image_url: cancha.image_url || '',
    });
    setShowCanchaModal(true);
  };

  const handleCanchaFormChange = (e) => {
    const { name, value, type, checked } = e.target;
    setCanchaForm({
      ...canchaForm,
      [name]: type === 'checkbox' ? checked : value,
    });
  };

  const handleSubmitCancha = async (e) => {
    e.preventDefault();

    const payload = {
      ...canchaForm,
      price: parseFloat(canchaForm.price),
      capacity: parseInt(canchaForm.capacity),
    };

    try {
      if (editingCancha) {
        await canchaService.updateCancha(editingCancha.id, payload, token);
      } else {
        await canchaService.createCancha(payload, token);
      }
      setShowCanchaModal(false);
      fetchCanchas();
    } catch (err) {
      alert('Error al guardar la cancha: ' + (err.response?.data?.message || err.message));
    }
  };

  const handleDeleteCancha = async (id) => {
    if (!window.confirm('¿Estás seguro de eliminar esta cancha?')) return;

    try {
      await canchaService.deleteCancha(id, token);
      fetchCanchas();
    } catch (err) {
      alert('Error al eliminar la cancha');
    }
  };

  const handleDeleteReserva = async (id) => {
    if (!window.confirm('¿Estás seguro de cancelar esta reserva?')) return;

    try {
      await reservaService.cancelReserva(id, token);
      fetchReservas();
    } catch (err) {
      alert('Error al cancelar la reserva');
    }
  };

  return (
    <div style={styles.container}>
      <h1 style={styles.title}>Panel de Administración</h1>

      {/* Tabs */}
      <div style={styles.tabs}>
        <button
          onClick={() => setActiveTab('canchas')}
          style={{
            ...styles.tab,
            ...(activeTab === 'canchas' ? styles.tabActive : {}),
          }}
        >
          Canchas ({canchas.length})
        </button>
        <button
          onClick={() => setActiveTab('reservas')}
          style={{
            ...styles.tab,
            ...(activeTab === 'reservas' ? styles.tabActive : {}),
          }}
        >
          Reservas ({reservas.length})
        </button>
      </div>

      {/* Contenido de Canchas */}
      {activeTab === 'canchas' && (
        <div style={styles.tabContent}>
          <div style={styles.header}>
            <h2 style={styles.subtitle}>Gestión de Canchas</h2>
            <button onClick={handleCreateCancha} style={styles.createBtn}>
              + Nueva Cancha
            </button>
          </div>

          {canchasLoading ? (
            <div style={styles.loading}>Cargando...</div>
          ) : (
            <div style={styles.table}>
              <table style={styles.tableElement}>
                <thead>
                  <tr style={styles.tableHeader}>
                    <th style={styles.th}>Nombre</th>
                    <th style={styles.th}>Tipo</th>
                    <th style={styles.th}>Ubicación</th>
                    <th style={styles.th}>Precio</th>
                    <th style={styles.th}>Capacidad</th>
                    <th style={styles.th}>Estado</th>
                    <th style={styles.th}>Acciones</th>
                  </tr>
                </thead>
                <tbody>
                  {canchas.map((cancha) => (
                    <tr key={cancha.id} style={styles.tableRow}>
                      <td style={styles.td}>{cancha.name}</td>
                      <td style={styles.td}>
                        <span style={styles.typeBadge}>{cancha.type}</span>
                      </td>
                      <td style={styles.td}>{cancha.location}</td>
                      <td style={styles.td}>${cancha.price}</td>
                      <td style={styles.td}>{cancha.capacity}</td>
                      <td style={styles.td}>
                        <span
                          style={{
                            ...styles.statusBadge,
                            backgroundColor: cancha.available ? '#27ae60' : '#e74c3c',
                          }}
                        >
                          {cancha.available ? 'Disponible' : 'No disponible'}
                        </span>
                      </td>
                      <td style={styles.td}>
                        <button
                          onClick={() => handleEditCancha(cancha)}
                          style={styles.editBtn}
                        >
                          Editar
                        </button>
                        <button
                          onClick={() => handleDeleteCancha(cancha.id)}
                          style={styles.deleteBtn}
                        >
                          Eliminar
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      )}

      {/* Contenido de Reservas */}
      {activeTab === 'reservas' && (
        <div style={styles.tabContent}>
          <h2 style={styles.subtitle}>Gestión de Reservas</h2>

          {reservasLoading ? (
            <div style={styles.loading}>Cargando...</div>
          ) : (
            <div style={styles.table}>
              <table style={styles.tableElement}>
                <thead>
                  <tr style={styles.tableHeader}>
                    <th style={styles.th}>Usuario</th>
                    <th style={styles.th}>Cancha</th>
                    <th style={styles.th}>Fecha</th>
                    <th style={styles.th}>Horario</th>
                    <th style={styles.th}>Precio</th>
                    <th style={styles.th}>Estado</th>
                    <th style={styles.th}>Acciones</th>
                  </tr>
                </thead>
                <tbody>
                  {reservas.map((reserva) => (
                    <tr key={reserva.id} style={styles.tableRow}>
                      <td style={styles.td}>{reserva.user_name}</td>
                      <td style={styles.td}>{reserva.cancha_name}</td>
                      <td style={styles.td}>{reserva.date}</td>
                      <td style={styles.td}>
                        {reserva.start_time} - {reserva.end_time}
                      </td>
                      <td style={styles.td}>${reserva.total_price}</td>
                      <td style={styles.td}>
                        <span
                          style={{
                            ...styles.statusBadge,
                            backgroundColor: getStatusColor(reserva.status),
                          }}
                        >
                          {reserva.status}
                        </span>
                      </td>
                      <td style={styles.td}>
                        {reserva.status !== 'cancelled' && (
                          <button
                            onClick={() => handleDeleteReserva(reserva.id)}
                            style={styles.deleteBtn}
                          >
                            Cancelar
                          </button>
                        )}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      )}

      {/* Modal para Crear/Editar Cancha */}
      {showCanchaModal && (
        <div style={styles.modalOverlay}>
          <div style={styles.modal}>
            <h2 style={styles.modalTitle}>
              {editingCancha ? 'Editar Cancha' : 'Nueva Cancha'}
            </h2>

            <form onSubmit={handleSubmitCancha} style={styles.form}>
              <div style={styles.formRow}>
                <div style={styles.formGroup}>
                  <label style={styles.label}>Nombre</label>
                  <input
                    type="text"
                    name="name"
                    value={canchaForm.name}
                    onChange={handleCanchaFormChange}
                    required
                    style={styles.input}
                  />
                </div>

                <div style={styles.formGroup}>
                  <label style={styles.label}>Tipo</label>
                  <select
                    name="type"
                    value={canchaForm.type}
                    onChange={handleCanchaFormChange}
                    required
                    style={styles.input}
                  >
                    <option value="futbol">Fútbol</option>
                    <option value="tenis">Tenis</option>
                    <option value="basquet">Básquet</option>
                    <option value="paddle">Paddle</option>
                    <option value="voley">Voley</option>
                  </select>
                </div>
              </div>

              <div style={styles.formGroup}>
                <label style={styles.label}>Descripción</label>
                <textarea
                  name="description"
                  value={canchaForm.description}
                  onChange={handleCanchaFormChange}
                  required
                  rows="3"
                  style={styles.textarea}
                />
              </div>

              <div style={styles.formRow}>
                <div style={styles.formGroup}>
                  <label style={styles.label}>Ubicación</label>
                  <input
                    type="text"
                    name="location"
                    value={canchaForm.location}
                    onChange={handleCanchaFormChange}
                    required
                    style={styles.input}
                  />
                </div>

                <div style={styles.formGroup}>
                  <label style={styles.label}>Dirección</label>
                  <input
                    type="text"
                    name="address"
                    value={canchaForm.address}
                    onChange={handleCanchaFormChange}
                    required
                    style={styles.input}
                  />
                </div>
              </div>

              <div style={styles.formRow}>
                <div style={styles.formGroup}>
                  <label style={styles.label}>Precio (por hora)</label>
                  <input
                    type="number"
                    name="price"
                    value={canchaForm.price}
                    onChange={handleCanchaFormChange}
                    required
                    min="0"
                    step="0.01"
                    style={styles.input}
                  />
                </div>

                <div style={styles.formGroup}>
                  <label style={styles.label}>Capacidad</label>
                  <input
                    type="number"
                    name="capacity"
                    value={canchaForm.capacity}
                    onChange={handleCanchaFormChange}
                    required
                    min="1"
                    style={styles.input}
                  />
                </div>
              </div>

              <div style={styles.formGroup}>
                <label style={styles.label}>URL de Imagen</label>
                <input
                  type="url"
                  name="image_url"
                  value={canchaForm.image_url}
                  onChange={handleCanchaFormChange}
                  style={styles.input}
                  placeholder="https://ejemplo.com/imagen.jpg"
                />
              </div>

              <div style={styles.checkboxGroup}>
                <label style={styles.checkboxLabel}>
                  <input
                    type="checkbox"
                    name="available"
                    checked={canchaForm.available}
                    onChange={handleCanchaFormChange}
                    style={styles.checkbox}
                  />
                  Disponible
                </label>
              </div>

              <div style={styles.modalActions}>
                <button
                  type="button"
                  onClick={() => setShowCanchaModal(false)}
                  style={styles.cancelModalBtn}
                >
                  Cancelar
                </button>
                <button type="submit" style={styles.submitBtn}>
                  {editingCancha ? 'Actualizar' : 'Crear'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};

const getStatusColor = (status) => {
  const colors = {
    confirmed: '#27ae60',
    pending: '#f39c12',
    cancelled: '#e74c3c',
  };
  return colors[status] || '#95a5a6';
};

const styles = {
  container: {
    maxWidth: '1400px',
    margin: '0 auto',
    padding: '2rem 1rem',
  },
  title: {
    fontSize: '2rem',
    marginBottom: '2rem',
    color: '#2c3e50',
  },
  tabs: {
    display: 'flex',
    gap: '0.5rem',
    marginBottom: '2rem',
    borderBottom: '2px solid #ecf0f1',
  },
  tab: {
    padding: '0.75rem 1.5rem',
    backgroundColor: 'transparent',
    border: 'none',
    borderBottom: '2px solid transparent',
    cursor: 'pointer',
    fontSize: '1rem',
    color: '#7f8c8d',
    transition: 'all 0.3s',
  },
  tabActive: {
    color: '#3498db',
    borderBottomColor: '#3498db',
    fontWeight: '500',
  },
  tabContent: {
    backgroundColor: '#fff',
    padding: '2rem',
    borderRadius: '8px',
    boxShadow: '0 2px 8px rgba(0,0,0,0.1)',
  },
  header: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: '1.5rem',
  },
  subtitle: {
    fontSize: '1.5rem',
    color: '#2c3e50',
    margin: 0,
  },
  createBtn: {
    backgroundColor: '#27ae60',
    color: '#fff',
    padding: '0.75rem 1.5rem',
    border: 'none',
    borderRadius: '4px',
    cursor: 'pointer',
    fontSize: '1rem',
    fontWeight: '500',
  },
  loading: {
    textAlign: 'center',
    padding: '2rem',
    color: '#7f8c8d',
  },
  table: {
    overflowX: 'auto',
  },
  tableElement: {
    width: '100%',
    borderCollapse: 'collapse',
  },
  tableHeader: {
    backgroundColor: '#f8f9fa',
  },
  th: {
    padding: '1rem',
    textAlign: 'left',
    fontWeight: '500',
    color: '#2c3e50',
    borderBottom: '2px solid #ecf0f1',
  },
  tableRow: {
    borderBottom: '1px solid #ecf0f1',
  },
  td: {
    padding: '1rem',
    color: '#2c3e50',
  },
  typeBadge: {
    backgroundColor: '#3498db',
    color: '#fff',
    padding: '0.25rem 0.75rem',
    borderRadius: '12px',
    fontSize: '0.85rem',
    textTransform: 'capitalize',
  },
  statusBadge: {
    color: '#fff',
    padding: '0.25rem 0.75rem',
    borderRadius: '12px',
    fontSize: '0.85rem',
    textTransform: 'capitalize',
  },
  editBtn: {
    backgroundColor: '#3498db',
    color: '#fff',
    padding: '0.5rem 1rem',
    border: 'none',
    borderRadius: '4px',
    cursor: 'pointer',
    marginRight: '0.5rem',
    fontSize: '0.85rem',
  },
  deleteBtn: {
    backgroundColor: '#e74c3c',
    color: '#fff',
    padding: '0.5rem 1rem',
    border: 'none',
    borderRadius: '4px',
    cursor: 'pointer',
    fontSize: '0.85rem',
  },
  modalOverlay: {
    position: 'fixed',
    top: 0,
    left: 0,
    right: 0,
    bottom: 0,
    backgroundColor: 'rgba(0,0,0,0.5)',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    zIndex: 1000,
  },
  modal: {
    backgroundColor: '#fff',
    padding: '2rem',
    borderRadius: '8px',
    maxWidth: '600px',
    width: '90%',
    maxHeight: '90vh',
    overflowY: 'auto',
  },
  modalTitle: {
    fontSize: '1.5rem',
    marginBottom: '1.5rem',
    color: '#2c3e50',
  },
  form: {
    display: 'flex',
    flexDirection: 'column',
    gap: '1rem',
  },
  formRow: {
    display: 'grid',
    gridTemplateColumns: '1fr 1fr',
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
    fontSize: '0.9rem',
  },
  input: {
    padding: '0.75rem',
    border: '1px solid #ddd',
    borderRadius: '4px',
    fontSize: '1rem',
  },
  textarea: {
    padding: '0.75rem',
    border: '1px solid #ddd',
    borderRadius: '4px',
    fontSize: '1rem',
    fontFamily: 'inherit',
    resize: 'vertical',
  },
  checkboxGroup: {
    display: 'flex',
    alignItems: 'center',
  },
  checkboxLabel: {
    display: 'flex',
    alignItems: 'center',
    gap: '0.5rem',
    cursor: 'pointer',
    fontSize: '0.95rem',
    color: '#2c3e50',
  },
  checkbox: {
    width: '18px',
    height: '18px',
    cursor: 'pointer',
  },
  modalActions: {
    display: 'flex',
    gap: '1rem',
    justifyContent: 'flex-end',
    marginTop: '1rem',
  },
  cancelModalBtn: {
    backgroundColor: '#95a5a6',
    color: '#fff',
    padding: '0.75rem 1.5rem',
    border: 'none',
    borderRadius: '4px',
    cursor: 'pointer',
    fontSize: '1rem',
  },
  submitBtn: {
    backgroundColor: '#27ae60',
    color: '#fff',
    padding: '0.75rem 1.5rem',
    border: 'none',
    borderRadius: '4px',
    cursor: 'pointer',
    fontSize: '1rem',
  },
};

export default Admin;