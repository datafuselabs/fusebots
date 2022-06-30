FROM golang:1.18-alpine as builder
COPY . /build/
RUN cd /build && go build -o fusebots ./cmd/

FROM alpine:latest
WORKDIR /app
COPY .github/ /app/.github/
COPY --from=builder /build/fusebots /bin/fusebots
ENTRYPOINT ["/bin/fusebots"]
