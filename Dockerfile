FROM golang:1.24.0 as builder

WORKDIR /build

ARG goproxy=https://goproxy.cn,direct
ENV GOPROXY ${goproxy}

COPY . .

RUN CGO_ENABLED=0 go build -o ks-mcp-server cmd/main.go

FROM alpine:3.21.3

COPY --from=builder /build/ks-mcp-server /usr/local/bin/ks-mcp-server

ENTRYPOINT ["mcp-server"]
