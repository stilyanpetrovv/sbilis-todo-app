# Basic todo app with golang

A simple web-based Todo List application built using Go and SQLite.

## Running the App Locally

```bash
git clone https://github.com/stilyanpetrovv/sbilis-todo-app.git
cd sbilis-todo-app
go mod tidy
go run main.go
```

## Or Build and Run it with Docker

```bash
docker build -t sbilis-todo-app .
docker run -p 8080:8080 sbilis-todo-app
```

Access the app at: http://localhost:8080


This focuses on just the essentials to get it up and running.
