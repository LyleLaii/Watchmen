FROM golang:1.14-alpine3.12 AS builder

WORKDIR /root/watchmen
COPY . .
#ARG GOPROXY="https://mirrors.aliyun.com/goproxy/"
RUN go env -w GOPROXY=https://goproxy.cn,direct
RUN CGO_ENABLED=0 GOOS=linux go build -a -trimpath -o bin/watchmen cmd/main.go

FROM debian:stretch-slim

WORKDIR /

COPY --from=builder /root/watchmen/bin /usr/local/bin

CMD ["watchmen"]