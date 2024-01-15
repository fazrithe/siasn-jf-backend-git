ARG BUILDER_IMAGE=golang:1.17-bullseye
ARG BASE_IMAGE=harbor.bkn.go.id/bkn/siasn-runner:latest

FROM $BUILDER_IMAGE as builder

COPY . /root/go/src/app/

ARG BUILD_VERSION
ARG GOPROXY
ARG GOSUMDB=sum.golang.org

WORKDIR /root/go/src/app

ENV PATH="${PATH}:/usr/local/go/bin"
ENV BUILD_VERSION=$BUILD_VERSION
ENV GOPROXY=$GOPROXY
ENV GOSUMDB=$GOSUMDB

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -v -ldflags "-X main.version=$(BUILD_VERSION)" -installsuffix cgo -o app .

FROM $BASE_IMAGE

WORKDIR /usr/app

COPY --from=builder /root/go/src/app/app /usr/app/app

# Included in the runner base image is siasn-docx script.
ENV SIASN_JF_SIASN_DOCX_CMD=siasn-docx

LABEL authors="Sergio Ryan <sergioryan@potatobeans.id>"

# PotatoBeans Co. adheres to OCI image specification.
LABEL org.opencontainers.image.authors="Sergio Ryan <sergioryan@potatobeans.id>"
LABEL org.opencontainers.image.title="siasn-jf-backend"
LABEL org.opencontainers.image.description="SIASN Manajemen JF Backend."
LABEL org.opencontainers.image.vendor="Institut Teknologi Bandung (ITB)"

ENTRYPOINT ["/usr/app/app"]
