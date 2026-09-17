FROM golang:1.25-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /servidor ./cmd/api

FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata \
 && adduser -D -u 10001 tienda
COPY --from=build /servidor /servidor
USER tienda
EXPOSE 8080
ENTRYPOINT ["/servidor"]
