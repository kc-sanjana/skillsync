FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /bin/api ./cmd/api

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
COPY --from=builder /bin/api /bin/api
COPY migrations/ /app/migrations/
WORKDIR /app
EXPOSE 8080
CMD ["/bin/api"]
