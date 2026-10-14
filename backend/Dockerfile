# BUILD STAGE
FROM golang:1.26.1-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/api ./cmd/api/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/seed ./cmd/seed/main.go

# RUN STAGE
FROM gcr.io/distroless/static-debian13
WORKDIR /app
COPY --from=builder /app/api .
COPY --from=builder /app/seed .
EXPOSE 3000
CMD [ "./api" ]
