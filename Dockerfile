FROM golang:latest AS builder

WORKDIR /app
COPY . .
RUN go mod download
RUN make build

FROM scratch

WORKDIR /app
ENV HOME /app

COPY --from=builder /app/bin/shurl ./bin
COPY --from=builder /app/config/docker/config.yml ./bin
COPY --from=builder /app/web/static ./web/static
COPY --from=builder /app/web/templates ./web/templates

EXPOSE 8443
#CMD ["bin/shurl"]
