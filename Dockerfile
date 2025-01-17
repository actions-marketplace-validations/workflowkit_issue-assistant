FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install git for private dependencies
RUN apk add --no-cache git

# Copy go.mod and go.sum first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the code
COPY . .

# Build the application
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /bin/issue-assistant

# Use a smaller base image for the final image
FROM gcr.io/distroless/static:nonroot

COPY --from=builder /bin/issue-assistant /bin/issue-assistant

ENTRYPOINT ["/bin/issue-assistant"] 