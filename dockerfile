# First stage: Build the Go app
FROM golang:1.20-bullseye as builder

# Set necessary environment variables
ENV CGO_ENABLED=1 GOOS=linux GOARCH=amd64

# Install SQLite development libraries
RUN apt-get update && apt-get install -y gcc libc-dev sqlite3 libsqlite3-dev && rm -rf /var/lib/apt/lists/*

# Set the working directory for the application
WORKDIR /app

# Copy the application source code and other necessary files
COPY . .

# Download dependencies and build the application
RUN go mod tidy && go build -o todo-app

# Second stage: Create a lightweight image with only the necessary runtime components
FROM gcr.io/distroless/base-debian10

# Set the working directory in the final container
WORKDIR /app

# Copy the compiled binary from the builder stage
COPY --from=builder /app/todo-app /app/todo-app

# Copy the templates and static assets
COPY --from=builder /app/templates /app/templates
COPY --from=builder /app/static /app/static

# Expose the application port
EXPOSE 8080

# Run the application
CMD ["/app/todo-app"]

