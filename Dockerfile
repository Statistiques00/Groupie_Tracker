FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY . .
RUN go build -o groupie-tracker .

FROM alpine:3.20

WORKDIR /app
COPY --from=builder /app/groupie-tracker .
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/static ./static

EXPOSE 8080

CMD ["./groupie-tracker", "-addr", ":8080"]