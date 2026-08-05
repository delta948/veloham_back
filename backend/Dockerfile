FROM golang:1.23-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/api ./cmd/api

FROM alpine:3.20
WORKDIR /app
RUN mkdir -p uploads
COPY --from=build /app/api /app/api
EXPOSE 8080
CMD ["/app/api"]
