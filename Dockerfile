FROM node:20-alpine AS frontend
WORKDIR /app
COPY web/package*.json ./web/
WORKDIR /app/web
RUN npm install
WORKDIR /app
COPY . .
WORKDIR /app/web
RUN npm run build

FROM golang:1.26-alpine AS backend
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend /app/internal/httpapi/webdist ./internal/httpapi/webdist
RUN go build -o /out/server ./cmd/server

FROM alpine:3.22
WORKDIR /app
RUN adduser -D -u 10001 appuser
COPY --from=backend /out/server /app/server
RUN mkdir -p /app/data && chown -R appuser:appuser /app
USER appuser
EXPOSE 8080
ENV RATI_ADDR=:8080
CMD ["/app/server"]
