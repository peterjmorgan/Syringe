FROM --platform=linux/amd64 golang:1.18-alpine

# CLI_VER specifies the Phylum CLI version to install in the image.
# Values should be provided in a format acceptable to the `phylum-init` script.
# When not defined, the value will default to `latest`.
# ARG CLI_VER

LABEL maintainer="Phylum, Inc. <engineering@phylum.io>"

WORKDIR /app

# Install go depdendencies
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

RUN go build -v -o /app/Syringe
RUN set -eux; \
    apk add --update --no-cache curl git minisign bash; \
    #phylum-init --phylum-release ${CLI_VER:-latest}; \
    curl -O "https://raw.githubusercontent.com/phylum-dev/cli/main/scripts/phylum-init.sh"; \
    chmod u+x phylum-init.sh; \
    ./phylum-init.sh -y

ENV PATH="/root/.local/bin:${PATH}"

ENTRYPOINT ["/app/Syringe"]
CMD ["--help"]
# CMD ["/bin/bash"]
# CMD ["phylum-ci"] \
