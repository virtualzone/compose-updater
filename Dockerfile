ARG UBUNTU_VERSION=21.10
ARG GO_VERSION=1.17
ARG COMPOSE_VERSION=2.2.2

FROM amd64/golang:${GO_VERSION} AS builder
RUN export GOBIN=$HOME/work/bin
WORKDIR /go/src/app
ADD src/ .
ADD src/go.mod .
ADD src/go.sum .
RUN go get -d -v ./...
RUN CGO_ENABLED=1 go build -o main .

FROM amd64/ubuntu:${UBUNTU_VERSION}
ARG COMPOSE_VERSION
ARG BUILD_DATE
ARG VCS_REF
LABEL org.label-schema.build-date=$BUILD_DATE \
        org.label-schema.name="Compose Updater" \
        org.label-schema.description="Automatically update your Docker Compose containers." \
        org.label-schema.vcs-ref=$VCS_REF \
        org.label-schema.vcs-url="https://github.com/virtualzone/compose-updater" \
        org.label-schema.schema-version="1.0"
RUN apt-get update && apt-get -y install \
    ca-certificates \
    curl \
    gnupg \
    lsb-release \
    curl
RUN curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg
RUN echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu \
  $(lsb_release -cs) stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null
RUN apt-get update && apt-get -y install docker-ce docker-ce-cli containerd.io
RUN curl -L "https://github.com/docker/compose/releases/download/v${COMPOSE_VERSION}/docker-compose-linux-x86_64" -o /usr/local/bin/docker-compose && \
    chmod +x /usr/local/bin/docker-compose
COPY --from=builder /go/src/app/main /usr/local/bin/docker-compose-watcher
CMD ["docker-compose-watcher", "-once=0", "-printSettings", "-cleanup"]