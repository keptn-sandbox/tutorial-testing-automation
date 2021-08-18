FROM golang:latest

# Set the Current Working Directory inside the container
WORKDIR /app

RUN GO111MODULE=on

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

COPY . .

# Run tests
CMD go test ./... -v