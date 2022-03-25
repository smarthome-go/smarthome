FROM amd64/debian:stable-slim

COPY ./app /app
COPY motd /etc/motd
COPY homescript /bin/
WORKDIR /app/

RUN export DEBIAN_FRONTEND=noninteractive && apt update && apt install -yq \
  tzdata \
  curl \
  ca-certificates \
  && apt autoclean && apt autoremove && rm -rf /var/lib/apt/lists/*

RUN echo '[ ! -z "$TERM" -a -r /etc/motd ] && cat /etc/motd' >> /etc/bash.bashrc

HEALTHCHECK --interval=300s --timeout=5s --start-period=5s CMD curl --fail http://localhost:80/health || exit 1
EXPOSE 80
ENTRYPOINT ["/app/smarthome"]