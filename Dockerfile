# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /src

RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /agent ./main.go

# Runtime stage
FROM alpine:3.24

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app
COPY --from=builder /agent /usr/local/bin/agent

# add kubectl binary
RUN apk add --no-cache curl
RUN curl -LO "https://dl.k8s.io/release/$(curl -Ls https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl" && chmod +x kubectl
COPY kubectl /usr/local/bin/kubectl

ENTRYPOINT ["/usr/local/bin/agent"]
