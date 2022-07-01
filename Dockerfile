FROM golang:1.18-alpine as builder
COPY . /build/
RUN cd /build && go build -o fusebots ./cmd/

FROM alpine:latest
RUN apk add --no-cache tzdata && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone && \
    apk del tzdata
WORKDIR /app
COPY .github/ /app/.github/
COPY --from=builder /build/fusebots /bin/fusebots
ENTRYPOINT ["/bin/fusebots"]
