FROM golang:latest AS builder

WORKDIR /app
COPY . .
RUN go mod download
RUN make build

FROM scratch

COPY --from=builder /app/bin/shurl /app/bin
COPY --from=builder /app/config/docker/config.yml /app/bin
COPY --from=builder /app/web /app/web

ENV HOME /app
WORKDIR /app

EXPOSE 8443
CMD ["/app/bin/shurl -config /app/bin/config.yml"]
