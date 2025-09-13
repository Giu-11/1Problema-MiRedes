FROM golang:1.25.0 AS builder

WORKDIR /app

COPY . .

RUN go mod tidy

RUN go build -o /app/bin/cliente ./cliente

FROM golang:1.25.0

WORKDIR /app

COPY --from=builder /app/bin/cliente .

CMD ["./cliente", "-docker"]