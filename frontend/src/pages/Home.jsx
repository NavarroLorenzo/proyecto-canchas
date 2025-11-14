import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import searchService from '../services/searchService';
import canchaService from '../services/canchaService';

const Home = () => {
  const [canchas, setCanchas] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  // Filtros - TODO LO MANEJA EL BACKEND
  const [filters, setFilters] = useState({
    q: '',
    type: '',
    number: '',
    available: '',
    sort_by: 'name',
    sort_order: 'asc',
    page: 1,
    page_size: 12,
  });

  const [pagination, setPagination] = useState({
    total: 0,
    page: 1,
    page_size: 12,
    total_pages: 0,
  });

  const navigate = useNavigate();

  // Cargar canchas al iniciar y cuando cambian los filtros
  useEffect(() => {
    fetchCanchas();
  }, [filters]);

  const normalizeField = (value) => {
    if (Array.isArray(value)) {
      return value[0] ?? '';
    }
    return value ?? '';
  };

  const fetchCanchas = async () => {
    setLoading(true);
    setError('');

    try {
      // Preparar parámetros - eliminar valores vacíos
      const params = {};
      Object.keys(filters).forEach(key => {
        if (filters[key] !== '' && filters[key] !== null) {
          params[key] = filters[key];
        }
      });

      // LLAMADA AL BACKEND - search-api hace TODO el filtrado
      const response = await searchService.searchCanchas(params);
      let results = response.results || [];
      // Si el backend no filtra por number, aplicamos filtro en cliente como respaldo
      if (params.number) {
        results = results.filter(c => String(c.number) === String(params.number));
      }

      // Si search-api devolvió 0 resultados pero el usuario filtró por tipo,
      // intentar respaldo rápido: obtener todas las canchas desde canchas-api
      // y aplicar el filtro en cliente usando comparación sin acentos.
      const normalize = (s) => {
        if (!s) return '';
        return s.normalize('NFD').replace(/\p{M}/gu, '').toLowerCase().trim();
      };

      // Si no hay resultados y el usuario filtró por tipo o por disponibilidad,
      // intentar fallback obteniendo todas las canchas y filtrando en cliente.
      if ((results.length === 0) && (filters.type || filters.available === 'true')) {
        try {
          const allResp = await canchaService.getAllCanchas();
          let allResults = allResp.canchas || [];
          // Aplicar filtro por tipo si existe
          if (filters.type) {
            const wanted = normalize(filters.type);
            allResults = allResults.filter(c => normalize(c.type) === wanted);
          }
          // Aplicar filtro de disponibilidad si está activado
          if (filters.available === 'true') {
            allResults = allResults.filter(c => c.available === true);
          }
          // Además aplicar filtro por number si existe
          if (filters.number) {
            allResults = allResults.filter(c => String(c.number) === String(filters.number));
          }
          results = allResults;
          // Ajustar paginación local
          response.total = allResults.length;
          response.page = 1;
          response.page_size = allResults.length || filters.page_size;
          response.total_pages = 1;
        } catch (err) {
          // si falla el fallback, seguimos con resultados vacíos
        }
      }

      setCanchas(results);
      setPagination({
        total: response.total || 0,
        page: response.page || 1,
        page_size: response.page_size || 12,
        total_pages: response.total_pages || 0,
      });
    } catch (err) {
      console.error('Error fetching canchas:', err);
      // Si search-api no responde, intentar con canchas-api directamente
      try {
        const response = await canchaService.getAllCanchas();
        let results = response.canchas || [];

        // Aplicar filtros en cliente como respaldo
        if (filters.number) {
          results = results.filter(c => String(c.number) === String(filters.number));
        }
        if (filters.type) {
          results = results.filter(c => String(c.type).toLowerCase() === String(filters.type).toLowerCase());
        }
        if (filters.available === 'true') {
          results = results.filter(c => c.available === true);
        }
        if (filters.q) {
          const qLower = filters.q.toLowerCase();
          results = results.filter(c => ((c.name && c.name.toLowerCase().includes(qLower)) || (c.description && c.description.toLowerCase().includes(qLower))));
        }

        setCanchas(results);
        setPagination({
          total: results.length,
          page: 1,
          page_size: results.length || filters.page_size,
          total_pages: 1,
        });
      } catch (err2) {
        setError('Error al cargar las canchas');
      }
    } finally {
      setLoading(false);
    }
  };

  // Nota: eliminada la verificación por ID para evitar 404 en consola cuando el índice está desactualizado.

  const handleFilterChange = (e) => {
    const { name, value } = e.target;
    setFilters(prev => ({
      ...prev,
      [name]: value,
      page: 1, // Reset a página 1 cuando cambia un filtro
    }));
  };

  const handleSearch = (e) => {
    e.preventDefault();
    // Los filtros ya se aplicaron automáticamente con useEffect
  };

  const handlePageChange = (newPage) => {
    setFilters(prev => ({
      ...prev,
      page: newPage,
    }));
    window.scrollTo(0, 0);
  };

  const clearFilters = () => {
    setFilters({
      q: '',
      type: '',
      number: '',
      available: '',
      sort_by: 'name',
      sort_order: 'asc',
      page: 1,
      page_size: 12,
    });
  };

  return (
    <div style={styles.container}>
      <div style={styles.hero}>
        <h1 style={styles.heroTitle}>Reserva tu Cancha Deportiva</h1>
        <p style={styles.heroSubtitle}>
          Encuentra las mejores canchas de fútbol, tenis, básquet y más
        </p>
      </div>

      {/* FILTROS - Todo se envía al backend */}
      <div style={styles.filters}>
        <form onSubmit={handleSearch} style={styles.filterForm}>
          {/* Búsqueda general */}
          <input
            type="text"
            name="q"
            value={filters.q}
            onChange={handleFilterChange}
            placeholder="Buscar canchas..."
            style={styles.searchInput}
          />

          {/* Tipo de cancha */}
          <select
            name="type"
            value={filters.type}
            onChange={handleFilterChange}
            style={styles.select}
          >
            <option value="">Todos los tipos</option>
            <option value="futbol">Fútbol</option>
            <option value="tenis">Tenis</option>
            <option value="basquet">Básquet</option>
            <option value="paddle">Paddle</option>
            <option value="voley">Voley</option>
          </select>

          {/* Número de cancha (id) */}
          <input
            type="text"
            name="number"
            value={filters.number}
            onChange={handleFilterChange}
            placeholder="Número de cancha"
            style={styles.input}
          />

          {/* Solo disponibles */}
          <select
            name="available"
            value={filters.available}
            onChange={handleFilterChange}
            style={styles.select}
          >
            <option value="">Todas</option>
            <option value="true">Solo disponibles</option>
          </select>

          {/* Ordenar por */}
          <select
            name="sort_by"
            value={filters.sort_by}
            onChange={handleFilterChange}
            style={styles.select}
          >
            <option value="name">Nombre</option>
            <option value="price">Precio</option>
            <option value="capacity">Capacidad</option>
          </select>

          {/* Orden */}
          <select
            name="sort_order"
            value={filters.sort_order}
            onChange={handleFilterChange}
            style={styles.select}
          >
            <option value="asc">Ascendente</option>
            <option value="desc">Descendente</option>
          </select>

          <button type="button" onClick={clearFilters} style={styles.clearBtn}>
            Limpiar
          </button>
        </form>
      </div>

      {/* Resultados */}
      <div style={styles.content}>
        {loading ? (
          <div style={styles.loading}>Cargando canchas...</div>
        ) : error ? (
          <div style={styles.error}>{error}</div>
        ) : canchas.length === 0 ? (
          <div style={styles.noResults}>
            No se encontraron canchas con los filtros seleccionados
          </div>
        ) : (
          <>
            <div style={styles.resultsInfo}>
              Mostrando {canchas.length} de {pagination.total} canchas
            </div>

            <div style={styles.grid}>
              {canchas.map((cancha) => {
                const descriptionText = normalizeField(cancha.description);
                return (
                  <div key={cancha.id} style={styles.card}>
                    <div style={styles.cardImage}>
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

                    <div style={styles.cardContent}>
                      <h3 style={styles.cardTitle}>{cancha.name} <span style={styles.cardNumber}>#{cancha.number || cancha.id}</span></h3>

                      <div style={styles.cardInfo}>
                        <span style={styles.badge}>{cancha.type}</span>
                      </div>

                      <p style={styles.cardDescription}>
                        {descriptionText.substring(0, 100)}...
                      </p>

                      <div style={styles.cardFooter}>
                        <div style={styles.price}>${cancha.price}</div>
                        <button
                          onClick={() => navigate(`/cancha/${cancha.id}`)}
                          style={styles.detailsBtn}
                        >
                          Ver Detalles
                        </button>
                      </div>
                    </div>
                  </div>
                )
              })}
            </div>

            {/* Paginación */}
            {pagination.total_pages > 1 && (
              <div style={styles.pagination}>
                <button
                  onClick={() => handlePageChange(pagination.page - 1)}
                  disabled={pagination.page === 1}
                  style={{
                    ...styles.pageBtn,
                    ...(pagination.page === 1 ? styles.pageBtnDisabled : {}),
                  }}
                >
                  Anterior
                </button>

                <span style={styles.pageInfo}>
                  Página {pagination.page} de {pagination.total_pages}
                </span>

                <button
                  onClick={() => handlePageChange(pagination.page + 1)}
                  disabled={pagination.page === pagination.total_pages}
                  style={{
                    ...styles.pageBtn,
                    ...(pagination.page === pagination.total_pages ? styles.pageBtnDisabled : {}),
                  }}
                >
                  Siguiente
                </button>
              </div>
            )}
          </>
        )}
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
    minHeight: '100vh',
    backgroundColor: '#f5f5f5',
  },
  hero: {
    backgroundColor: '#3498db',
    color: '#fff',
    padding: '3rem 1rem',
    textAlign: 'center',
  },
  heroTitle: {
    fontSize: '2.5rem',
    marginBottom: '0.5rem',
  },
  heroSubtitle: {
    fontSize: '1.2rem',
    opacity: 0.9,
  },
  filters: {
    backgroundColor: '#fff',
    padding: '1.5rem',
    boxShadow: '0 2px 4px rgba(0,0,0,0.1)',
  },
  filterForm: {
    maxWidth: '1200px',
    margin: '0 auto',
    display: 'grid',
    gridTemplateColumns: 'repeat(auto-fit, minmax(150px, 1fr))',
    gap: '1rem',
    alignItems: 'center',
  },
  searchInput: {
    gridColumn: '1 / -1',
    padding: '0.75rem',
    border: '1px solid #ddd',
    borderRadius: '4px',
    fontSize: '1rem',
  },
  input: {
    padding: '0.75rem',
    border: '1px solid #ddd',
    borderRadius: '4px',
    fontSize: '0.9rem',
  },
  inputSmall: {
    padding: '0.75rem',
    border: '1px solid #ddd',
    borderRadius: '4px',
    fontSize: '0.9rem',
  },
  select: {
    padding: '0.75rem',
    border: '1px solid #ddd',
    borderRadius: '4px',
    fontSize: '0.9rem',
    backgroundColor: '#fff',
  },
  clearBtn: {
    padding: '0.75rem',
    backgroundColor: '#95a5a6',
    color: '#fff',
    border: 'none',
    borderRadius: '4px',
    cursor: 'pointer',
    fontSize: '0.9rem',
  },
  content: {
    maxWidth: '1200px',
    margin: '0 auto',
    padding: '2rem 1rem',
  },
  loading: {
    textAlign: 'center',
    padding: '3rem',
    fontSize: '1.2rem',
    color: '#7f8c8d',
  },
  error: {
    backgroundColor: '#fee',
    color: '#c00',
    padding: '1rem',
    borderRadius: '4px',
    textAlign: 'center',
  },
  noResults: {
    textAlign: 'center',
    padding: '3rem',
    fontSize: '1.2rem',
    color: '#7f8c8d',
  },
  resultsInfo: {
    marginBottom: '1rem',
    color: '#7f8c8d',
    fontSize: '0.9rem',
  },
  grid: {
    display: 'grid',
    gridTemplateColumns: 'repeat(auto-fill, minmax(280px, 1fr))',
    gap: '1.5rem',
  },
  card: {
    backgroundColor: '#fff',
    borderRadius: '8px',
    overflow: 'hidden',
    boxShadow: '0 2px 8px rgba(0,0,0,0.1)',
    transition: 'transform 0.2s',
    cursor: 'pointer',
  },
  cardImage: {
    height: '180px',
    backgroundColor: '#ecf0f1',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
  },
  image: {
    width: '100%',
    height: '100%',
    objectFit: 'cover',
  },
  imagePlaceholder: {
    fontSize: '4rem',
  },
  cardContent: {
    padding: '1rem',
  },
  cardTitle: {
    fontSize: '1.2rem',
    marginBottom: '0.5rem',
    color: '#2c3e50',
  },
  cardNumber: {
    fontSize: '0.85rem',
    color: '#7f8c8d',
    marginLeft: '0.5rem',
    fontWeight: '600',
  },
  cardInfo: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: '0.5rem',
  },
  badge: {
    backgroundColor: '#3498db',
    color: '#fff',
    padding: '0.25rem 0.5rem',
    borderRadius: '12px',
    fontSize: '0.8rem',
    textTransform: 'capitalize',
  },
  cardLocation: {
    fontSize: '0.85rem',
    color: '#7f8c8d',
  },
  cardDescription: {
    fontSize: '0.9rem',
    color: '#7f8c8d',
    marginBottom: '1rem',
    lineHeight: '1.4',
  },
  cardFooter: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
  },
  price: {
    fontSize: '1.3rem',
    fontWeight: 'bold',
    color: '#27ae60',
  },
  detailsBtn: {
    backgroundColor: '#3498db',
    color: '#fff',
    padding: '0.5rem 1rem',
    border: 'none',
    borderRadius: '4px',
    cursor: 'pointer',
    fontSize: '0.9rem',
  },
  pagination: {
    display: 'flex',
    justifyContent: 'center',
    alignItems: 'center',
    gap: '1rem',
    marginTop: '2rem',
  },
  pageBtn: {
    padding: '0.5rem 1rem',
    backgroundColor: '#3498db',
    color: '#fff',
    border: 'none',
    borderRadius: '4px',
    cursor: 'pointer',
  },
  pageBtnDisabled: {
    backgroundColor: '#bdc3c7',
    cursor: 'not-allowed',
  },
  pageInfo: {
    color: '#7f8c8d',
  },
};

export default Home;
