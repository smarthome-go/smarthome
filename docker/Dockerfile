FROM amd64/debian:stable-slim

LABEL author="MikMuellerDev"
LABEL version="0.0.10-beta"

RUN useradd -M smarthome

COPY ./app /app

COPY motd /etc/motd
RUN echo '[ ! -z "$TERM" -a -r /etc/motd ] && cat /etc/motd' >> /etc/bash.bashrc

ADD https://github.com/MikMuellerDev/homescript-cli/releases/download/v0.3.1-beta/homescript_linux_amd64.tar.gz /tmp/homescript

WORKDIR /usr/bin
RUN tar -xvf /tmp/homescript
RUN chmod +x /usr/bin/homescript

RUN chown -R smarthome:smarthome /app
WORKDIR /app/

RUN export DEBIAN_FRONTEND=noninteractive && apt update && apt install -yq \
  # tzdata \
  curl \
  && apt autoclean && apt autoremove && rm -rf /var/lib/apt/lists/*

HEALTHCHECK --interval=300s --timeout=5s --start-period=5s CMD curl --fail http://localhost:80/health || exit 1

EXPOSE 80/tcp

USER smarthome:smarthome

ENTRYPOINT ["/app/smarthome"]