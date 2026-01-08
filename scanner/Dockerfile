# Scanner image with Syft and Grype
FROM alpine:latest

# Pin versions for reproducibility and vulnerability scanning
ARG SYFT_VERSION=v1.40.0
ARG GRYPE_VERSION=v0.104.4

# Install dependencies
RUN apk add --no-cache \
    bash \
    curl \
    jq \
    ca-certificates

# Install Syft with pinned version
RUN curl -sSfL https://raw.githubusercontent.com/anchore/syft/main/install.sh | sh -s -- -b /usr/local/bin ${SYFT_VERSION}

# Install Grype with pinned version
RUN curl -sSfL https://raw.githubusercontent.com/anchore/grype/main/install.sh | sh -s -- -b /usr/local/bin ${GRYPE_VERSION}

# Create metadata file documenting installed tool versions for vulnerability scanning
RUN echo "# Scanner Dependencies" > /scanner-dependencies.txt && \
    echo "syft_version=${SYFT_VERSION}" >> /scanner-dependencies.txt && \
    echo "grype_version=${GRYPE_VERSION}" >> /scanner-dependencies.txt && \
    echo "alpine_version=$(cat /etc/alpine-release)" >> /scanner-dependencies.txt && \
    echo "bash_version=$(bash --version | head -1)" >> /scanner-dependencies.txt && \
    echo "curl_version=$(curl --version | head -1)" >> /scanner-dependencies.txt && \
    echo "jq_version=$(jq --version)" >> /scanner-dependencies.txt

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
