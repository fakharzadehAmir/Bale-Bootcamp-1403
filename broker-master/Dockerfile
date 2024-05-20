# Use the latest Go image as the base image
FROM golang:1.20-alpine3.18

# Set the GOPATH environment variable
ENV GOPATH=/go

# Set the working directory inside the container
WORKDIR /go/src/broker

# Copy the source code into the container
COPY . .

# Install the application
RUN go install -v ./...

# Expose the application port
EXPOSE 8080

# Command to run the executable
CMD ["/go/bin/therealbroker"]
