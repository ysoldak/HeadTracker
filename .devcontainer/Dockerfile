ARG base_tag=ubuntu-24.04
ARG base_img=mcr.microsoft.com/devcontainers/base:dev-${base_tag}

FROM --platform=linux/amd64 ${base_img}

RUN \
  apt-get update --fix-missing && apt-get -y upgrade

RUN \
  wget https://go.dev/dl/go1.23.5.linux-amd64.tar.gz && \
  tar -C /usr/local -xzf go1.23.5.linux-amd64.tar.gz && \
  rm go1.23.5.linux-amd64.tar.gz && \
  echo 'export PATH=$PATH:/usr/local/go/bin' >> /home/vscode/.bashrc

RUN \
  wget https://github.com/tinygo-org/tinygo/releases/download/v0.35.0/tinygo_0.35.0_amd64.deb && \
  dpkg -i tinygo_0.35.0_amd64.deb && \
  rm tinygo_0.35.0_amd64.deb
