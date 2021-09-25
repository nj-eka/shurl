FROM golang:latest AS builder

WORKDIR /app
COPY . .
RUN go mod download
RUN curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.42.1
RUN make build

FROM scratch
WORKDIR /app
COPY --from=builder /app/build/shurl .
COPY --from=builder /app/web /app/web
COPY --from=builder /app/config/pro .
EXPOSE 8443
CMD ["./shurl"]
