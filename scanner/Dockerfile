# Scanner image with Syft and Grype
FROM alpine:latest

# Install dependencies
RUN apk add --no-cache \
    bash \
    curl \
    jq \
    ca-certificates

# Install Syft
RUN curl -sSfL https://raw.githubusercontent.com/anchore/syft/main/install.sh | sh -s -- -b /usr/local/bin

# Install Grype
RUN curl -sSfL https://raw.githubusercontent.com/anchore/grype/main/install.sh | sh -s -- -b /usr/local/bin

# Create non-root user
RUN addgroup -g 1000 scanner && \
    adduser -D -u 1000 -G scanner scanner

# Copy scanner script
COPY scanner.sh /scanner.sh
RUN chmod +x /scanner.sh

# Set working directory
WORKDIR /home/scanner

# Change ownership to non-root user
RUN chown -R scanner:scanner /home/scanner

# Switch to non-root user
USER scanner

ENTRYPOINT ["/scanner.sh"]
