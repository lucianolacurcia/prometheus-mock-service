FROM golang:1.18 as builder
WORKDIR /app
COPY go.mod .
COPY go.sum .
COPY *.go .
# COPY ./metrics.yaml ./config.yml
RUN go mod tidy
RUN CGO_ENABLED=0 go build -o /service

FROM alpine:latest
COPY --from=builder /service .
EXPOSE 5000
CMD ["/service"]
