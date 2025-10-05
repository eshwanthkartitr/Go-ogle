# syntax=docker/dockerfile:1
ARG SERVICE=searchapi

FROM golang:1.24 AS builder
ARG SERVICE
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/service ./cmd/${SERVICE}

FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=builder /out/service /usr/bin/service
COPY testdata /app/testdata
ENTRYPOINT ["/usr/bin/service"]
