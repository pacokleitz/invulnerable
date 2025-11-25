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

# Copy scanner script
COPY scanner.sh /scanner.sh
RUN chmod +x /scanner.sh

# Set working directory
WORKDIR /scan

ENTRYPOINT ["/scanner.sh"]
