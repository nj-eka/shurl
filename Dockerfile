FROM golang:latest AS builder

WORKDIR /app
COPY . .
RUN go mod download
RUN make build

FROM scratch
WORKDIR /bin
COPY --from=builder /app/bin/shurl .
COPY --from=builder /app/config/docker/config.yml .
COPY --from=builder /app/web ./web
EXPOSE 8443
CMD ["shurl"]
