FROM golang:1.17-alpine AS builder

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN go build -o dummy-webservice .

FROM alpine:latest  

WORKDIR /root/

COPY --from=builder /app/dummy-webservice .

RUN apk add --no-cache redis

ENV REDIS_ADDR=localhost:6379

EXPOSE 8080

# Command to run the executable
CMD ["./dummy-webservice"]
