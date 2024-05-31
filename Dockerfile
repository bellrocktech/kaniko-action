FROM gcr.io/kaniko-project/executor:v1.23.0-debug as executor

RUN wget -O /crane.tar.gz \
    https://github.com/google/go-containerregistry/releases/download/v0.13.0/go-containerregistry_Linux_x86_64.tar.gz && \
    tar -xvzf /crane.tar.gz crane -C /kaniko && \
    rm /crane.tar.gz

FROM golang:1.20-bullseye as go

COPY main.go /main.go

ENV CGO_ENABLED=0

WORKDIR /workspace

RUN go build -o /builder /main.go

FROM scratch

COPY --from=executor /kaniko /kaniko
COPY --from=executor /etc/nsswitch.conf /etc/nsswitch.conf
COPY --from=go /builder /builder

ENV HOME /root
ENV USER root
ENV SSL_CERT_DIR=/kaniko/ssl/certs
ENV DOCKER_CONFIG /kaniko/.docker/

CMD ["/builder"]
