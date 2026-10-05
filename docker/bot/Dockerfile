FROM golang:1.25.0-alpine AS builder

WORKDIR /app
RUN apk add --no-cache git ca-certificates
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o mojo -buildvcs=false .

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/mojo /mojo
COPY --from=builder /app/conf /conf
RUN chmod +x /mojo
CMD ["/mojo", "--run_bot=true"]
