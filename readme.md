# Plataforma de reservas de canchas

Proyecto académico de una aplicación para consultar canchas y gestionar reservas. Combina una interfaz en **React** con servicios en **Go** y una infraestructura local compuesta por **MySQL, MongoDB, RabbitMQ, Solr y Memcached**.

## Qué contiene

| Componente | Función |
| --- | --- |
| `frontend/` | Interfaz para buscar canchas, registrarse, iniciar sesión y consultar reservas. |
| `users-api/` | Usuarios y autenticación con JWT; persiste datos en MySQL. |
| `canchas-api/` | Gestión de canchas; utiliza MongoDB. |
| `reservas-api/` | Gestión de reservas; utiliza MongoDB. |
| `search-api/` | Búsqueda de canchas con Solr y caché con Memcached. |
| `docker-compose.yml` | Orquestación local de los servicios y sus dependencias. |

RabbitMQ conecta partes de la aplicación mediante eventos. El repositorio permite ver una arquitectura de varios servicios y el uso conjunto de bases relacionales y no relacionales.

## Ejecutarlo localmente

Requiere Docker con Compose. Desde la raíz del repositorio:

```bash
docker compose up --build
```

El archivo [`docker-compose.yml`](docker-compose.yml) expone el frontend en `http://localhost:5173` y las API en los puertos `8080` a `8083`. Los valores de base de datos y JWT definidos allí son **solo para desarrollo local**; hay que reemplazarlos antes de cualquier despliegue público.

Para detener el entorno:

```bash
docker compose down
```

Este README describe los componentes presentes en el repositorio. El proyecto es una entrega académica, no un servicio alojado ni una instalación lista para producción.
