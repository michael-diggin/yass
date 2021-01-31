FROM golang:1.14.6-alpine as builder

WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /bin/server github.com/michael-diggin/yass/server

# Adding the grpc_health_probe
RUN GRPC_HEALTH_PROBE_VERSION=v0.3.2 && \
    wget -qO/bin/grpc_health_probe https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/${GRPC_HEALTH_PROBE_VERSION}/grpc_health_probe-linux-amd64 && \
    chmod +x /bin/grpc_health_probe


# Final distroless image
FROM gcr.io/distroless/static
WORKDIR /root/
COPY --from=builder /bin/server .
COPY --from=builder /bin/grpc_health_probe ./grpc_health_probe
ENTRYPOINT ["./server"]