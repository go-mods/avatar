# Start from golang base image to build the server
FROM golang:1.20-alpine as builder

# Tools needed to compile
RUN apk update && apk add --no-cache git make

# Set the current working directory inside the container
WORKDIR /build

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies.
# Dependencies will be cached if the go.mod and the go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the working Directory inside the container
COPY . ./

# Build the Go app
RUN make build



# Start a new stage from scratch
FROM alpine:3.18

# Define working dir
WORKDIR /avatar

# Copy the Pre-built binary file from the previous stage.
COPY --from=builder /build/bin /avatar/

# Make avatar executable
RUN chmod +x avatar

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
CMD ["./avatar"]
