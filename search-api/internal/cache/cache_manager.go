package cache

import (
	"encoding/json"
	"log"
	"time"

	"search-api/config"
	"search-api/internal/dto"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/karlseguin/ccache/v3"
)

// Manager coordina la doble capa de caché: local (CCache) y distribuida (Memcached).
type Manager struct {
	local     *ccache.Cache[[]byte]
	remote    *memcache.Client
	localTTL  time.Duration
	remoteTTL time.Duration
}

// NewCacheManager construye la caché a partir de la configuración.
func NewCacheManager(cfg *config.Config) *Manager {
	if cfg == nil {
		return nil
	}

	manager := &Manager{}

	if cfg.LocalCacheSize > 0 {
		manager.local = ccache.New(ccache.Configure[[]byte]().MaxSize(int64(cfg.LocalCacheSize)))
		manager.localTTL = time.Duration(cfg.LocalCacheTTLMinutes) * time.Minute
	}

	if cfg.MemcachedURL != "" {
		manager.remote = memcache.New(cfg.MemcachedURL)
		manager.remoteTTL = time.Duration(cfg.MemcachedTTLMinutes) * time.Minute
		// Intentar un ping liviano para saber si está disponible.
		if err := manager.remote.Ping(); err != nil {
			log.Printf("[Cache] Memcached unreachable (%v), remote cache disabled", err)
			manager.remote = nil
		}
	}

	if manager.local == nil && manager.remote == nil {
		return nil
	}

	return manager
}

// GetSearchResponse busca primero en la caché local y luego en Memcached.
func (m *Manager) GetSearchResponse(key string) (*dto.SearchResponse, bool) {
	if m == nil {
		return nil, false
	}

	if data, ok := m.getFromLocal(key); ok {
		return decodeResponse(data)
	}

	if m.remote != nil {
		if item, err := m.remote.Get(key); err == nil && item != nil {
			if resp, ok := decodeResponse(item.Value); ok {
				// Recalentar capa local.
				m.saveLocal(key, item.Value)
				return resp, true
			}
		}
	}

	return nil, false
}

// SaveSearchResponse guarda el resultado en ambas capas.
func (m *Manager) SaveSearchResponse(key string, resp *dto.SearchResponse) {
	if m == nil || resp == nil {
		return
	}

	data, err := json.Marshal(resp)
	if err != nil {
		log.Printf("[Cache] Failed to marshal cache entry: %v", err)
		return
	}

	m.saveLocal(key, data)

	if m.remote != nil {
		exp := int32(m.remoteTTL.Seconds())
		if exp <= 0 {
			exp = 60
		}
		if err := m.remote.Set(&memcache.Item{Key: key, Value: data, Expiration: exp}); err != nil {
			log.Printf("[Cache] Failed to save entry in Memcached: %v", err)
		}
	}
}

// InvalidateAll limpia ambas capas. Se invoca cuando cambian las canchas.
func (m *Manager) InvalidateAll() {
	if m == nil {
		return
	}
	if m.local != nil {
		m.local.Clear()
	}
	if m.remote != nil {
		if err := m.remote.FlushAll(); err != nil {
			log.Printf("[Cache] Failed to flush Memcached: %v", err)
		}
	}
}

func (m *Manager) getFromLocal(key string) ([]byte, bool) {
	if m.local == nil {
		return nil, false
	}
	item := m.local.Get(key)
	if item == nil || item.Expired() {
		return nil, false
	}
	return item.Value(), true
}

func (m *Manager) saveLocal(key string, data []byte) {
	if m.local == nil {
		return
	}
	ttl := m.localTTL
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}
	m.local.Set(key, append([]byte(nil), data...), ttl)
}

func decodeResponse(data []byte) (*dto.SearchResponse, bool) {
	if len(data) == 0 {
		return nil, false
	}
	var resp dto.SearchResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		log.Printf("[Cache] Failed to decode cached response: %v", err)
		return nil, false
	}
	return &resp, true
}
