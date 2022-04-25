FROM alpine:3

LABEL author="MikMuellerDev"
LABEL version="0.0.25-beta-rc.2-fix"

COPY ./app /app

RUN addgroup smarthome && adduser \
    --disabled-password \
    --gecos "" \
    --home /app \
    --ingroup smarthome \
    --no-create-home \
    --uid 1000 \
    smarthome

RUN apk add curl bash && rm -rf /var/cache/apk/*

COPY motd /etc/motd
COPY .bashrc /app/.bashrc
ENV ENV=/app/.bashrc

RUN curl -L https://github.com/MikMuellerDev/homescript-cli/releases/download/v0.6.0-beta/homescript_linux_amd64.tar.gz -o /tmp/homescript.tar.gz
WORKDIR /bin
RUN tar -xvf /tmp/homescript.tar.gz && chmod +x /bin/homescript && rm -rf /tmp/homescript.tar.gz

RUN chown -R smarthome:smarthome /app
WORKDIR /app/

HEALTHCHECK --interval=300s --timeout=5s --start-period=5s CMD curl --fail http://localhost:80/health || exit 1

EXPOSE 80/tcp

USER smarthome:smarthome

ENV SMARTHOME_LOG_LEVEL=INFO

ENTRYPOINT ["/app/smarthome"]