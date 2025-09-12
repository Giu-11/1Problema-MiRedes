FROM golang:1.25.0 AS builder

WORKDIR /app

COPY . .

WORKDIR /app/servidor

RUN go mod tidy

RUN go build -o /app/bin/servidor .

FROM golang:1.25.0

WORKDIR /app

COPY --from=builder /app/bin/servidor .

EXPOSE 8080

CMD ["./servidor"]
