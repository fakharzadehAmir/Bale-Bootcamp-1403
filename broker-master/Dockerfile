# Use a minimal base image
FROM docker.arvancloud.ir/library/alpine:3.18

# Install necessary CA certificates for HTTPS connections
RUN apk add --no-cache ca-certificates

# Set the working directory inside the container
WORKDIR /app

# Copy the pre-built binary into the container
COPY therealbroker .

# Ensure the binary is executable
RUN chmod +x therealbroker

# Expose the application port (adjust if your application uses a different port)
EXPOSE 8080

# Command to run the executable
CMD ["/app/therealbroker"]
