FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o vram-guard .

FROM scratch
COPY --from=builder /app/vram-guard /vram-guard
EXPOSE 11434
ENTRYPOINT ["/vram-guard"]
