# Etapa 1: Build del frontend
FROM node:20-alpine AS builder
WORKDIR /app

# Copiar dependencias y archivo de entorno
COPY package*.json ./
COPY .env ./
RUN npm install

# Copiar todo el código fuente y construir
COPY . .
RUN npm run build

# Etapa 2: Servidor de producción con Nginx
FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html

# Configurar nginx
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 5173
CMD ["nginx", "-g", "daemon off;"]
