FROM golang:1.23-alpine AS builder

WORKDIR /build

RUN apk add --no-cache gcc musl-dev git

COPY go.mod ./
RUN go mod download || true

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -o picolcm ./cmd/picolcm

FROM alpine:3.19

RUN apk add --no-cache ca-certificates sqlite

WORKDIR /app

COPY --from=builder /build/picolcm .

RUN mkdir -p /app/data

COPY docker/config.json /app/data/config.json

ENV TZ=UTC

VOLUME /app/data

ENTRYPOINT ["./picolcm"]
