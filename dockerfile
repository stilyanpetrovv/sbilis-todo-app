# Start with an Ubuntu base image
FROM ubuntu:22.04

# Set environment variables
ENV GO_VERSION=1.23.3
ENV CGO_ENABLED=1
ENV GOOS=linux
ENV GOARCH=amd64

# Install dependencies, including Go and SQLite development libraries
RUN apt-get update && apt-get install -y \
    wget \
    build-essential \
    libsqlite3-dev \
    && rm -rf /var/lib/apt/lists/*

# Download and install Go
RUN wget https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz && \
    rm go${GO_VERSION}.linux-amd64.tar.gz

# Set Go path variables
ENV PATH="/usr/local/go/bin:${PATH}"

# Create a working directory for the app
WORKDIR /app

# Copy the application source code into the container
COPY . .

# Tidy and build the application with CGO enabled
RUN go mod tidy
RUN go build -o todo-app .

# Expose the app port
EXPOSE 8080

# Run the application
CMD ["./todo-app"]

