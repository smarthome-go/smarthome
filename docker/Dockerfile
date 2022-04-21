FROM alpine:3

LABEL author="MikMuellerDev"
LABEL version="0.0.22-beta"

COPY ./app /app

RUN addgroup smarthome

RUN adduser \
    --disabled-password \
    --gecos "" \
    --home /app \
    --ingroup smarthome \
    --no-create-home \
    --uid 1000 \
    smarthome

RUN apk add curl bash && rm -rf /var/cache/apk/*

COPY motd /etc/motd

COPY .ashrc /app/.ashrc
ENV ENV=/app/.ashrc

RUN curl -L https://github.com/MikMuellerDev/homescript-cli/releases/download/v0.5.2-beta/homescript_linux_amd64.tar.gz -o /tmp/homescript

WORKDIR /usr/bin
RUN tar -xvf /tmp/homescript
RUN chmod +x /usr/bin/homescript

RUN chown -R smarthome:smarthome /app
WORKDIR /app/

HEALTHCHECK --interval=300s --timeout=5s --start-period=5s CMD curl --fail http://localhost:80/health || exit 1

EXPOSE 80/tcp

USER smarthome:smarthome

ENV SMARTHOME_LOG_LEVEL=INFO

ENTRYPOINT ["/app/smarthome"]