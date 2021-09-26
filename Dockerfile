FROM golang:latest AS builder

WORKDIR /app
COPY . .
RUN go mod download
RUN make build

FROM scratch

WORKDIR /app
COPY --from=builder /app/bin/shurl /app
COPY --from=builder /app/config/docker/config.yml /app
COPY --from=builder /app/data /app/data
COPY --from=builder /app/web /app/web
ENV HOME /app

EXPOSE 8443
ENTRYPOINT ["/app/shurl"]
