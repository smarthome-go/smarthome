FROM amd64/debian:stable-slim

ADD ./app /app
ADD homescript /bin/
WORKDIR /app/

RUN export DEBIAN_FRONTEND=noninteractive && apt update && apt install -yq \
  tzdata \
  curl \
  ca-certificates \
  && apt autoclean && apt autoremove && rm -rf /var/lib/apt/lists/*

# HEALTHCHECK --interval=30s --timeout=5s --start-period=5s CMD curl --fail http://localhost:80/health || exit 1
EXPOSE 80
ENTRYPOINT ["/app/smarthome"]