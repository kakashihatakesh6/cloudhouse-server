FROM golang:1.20-alpine

WORKDIR /app
COPY . .

RUN go mod tidy

CMD ["go", "run", "main.go"]
