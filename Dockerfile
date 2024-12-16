# Step 1: Use the official Golang image to build the Go application
FROM golang:1.23.1-alpine AS builder

# Step 2: Set the Current Working Directory inside the container
WORKDIR /app

# Step 3: Copy the Go Modules manifests
COPY go.mod go.sum ./

# Step 4: Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod tidy

# Step 5: Copy the source code into the container
COPY . .

# Step 6: Build the Go app
RUN go build -o main .

# Step 7: Use a smaller base image to run the compiled Go app
FROM alpine:latest  

# Step 8: Install required dependencies (in this case, just a shell)
RUN apk --no-cache add ca-certificates

# Step 9: Set the Current Working Directory inside the container
WORKDIR /root/

# Step 10: Copy the compiled binary from the builder image
COPY --from=builder /app/main .

# Step 11: Copy the .env file into the container
COPY .env .env

# Step 11: Expose the port the app will run on
EXPOSE 8080

# Step 12: Command to run the application
CMD ["./main"]
