# Build the UI, then a static Go binary that serves the UI and the API.

FROM node:24.19.0-alpine AS frontend
WORKDIR /src/frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

FROM golang:1.27.4-alpine AS backend
WORKDIR /src
COPY backend/ ./
RUN CGO_ENABLED=0 go build -o /server ./cmd/server

FROM alpine:3.22
WORKDIR /app
RUN adduser -D -H -u 10001 app
COPY --from=backend /server /app/server
COPY --from=frontend /src/frontend/dist /app/static
USER app
ENV PORT=8080
ENV STATIC_DIR=/app/static
EXPOSE 8080
CMD ["/app/server"]
