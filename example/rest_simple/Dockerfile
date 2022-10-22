FROM golang:1.19 AS builder
WORKDIR /app
COPY ./ ./
RUN go mod tidy -compat=1.19
RUN go build -o app ./example/rest_simple

FROM ubuntu:latest AS app
WORKDIR /app
COPY --from=builder /app/app ./
CMD ["./app"]
