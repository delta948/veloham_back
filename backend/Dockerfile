FROM golang:1.26.5-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/api .

FROM alpine:3.20
WORKDIR /app
RUN addgroup -S veloham && adduser -S -G veloham veloham && mkdir -p uploads && chown -R veloham:veloham /app
COPY --from=build /app/api /app/api
USER veloham
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=3s --retries=3 CMD wget -qO- http://127.0.0.1:8080/healthz || exit 1
CMD ["/app/api"]
