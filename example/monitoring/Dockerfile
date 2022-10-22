FROM golang:1.19 AS builder
WORKDIR /app
COPY ./ ./
RUN cat go.mod
RUN go mod tidy -compat=1.19
RUN go build -o app ./example/monitoring

FROM ubuntu:latest AS app
WORKDIR /app
COPY --from=builder /app/app ./
CMD ["./app"]
