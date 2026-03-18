FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod ./
# go.sum is generated during build; run `go mod tidy` locally first for reproducible builds.
COPY . .
RUN go mod tidy && \
    CGO_ENABLED=0 GOOS=linux go build -o /go-backend .

FROM alpine:3.20
RUN addgroup -S app && adduser -S app -G app
COPY --from=builder /go-backend /go-backend
USER app
EXPOSE 8080
ENTRYPOINT ["/go-backend"]
