FROM alpine:3

LABEL author="MikMuellerDev"
LABEL version="0.4.2"

COPY ./cache /app

RUN addgroup smarthome && adduser \
    --disabled-password \
    --gecos "" \
    --home /app \
    --ingroup smarthome \
    --no-create-home \
    --uid 1000 \
    smarthome

RUN mkdir -p /app/.config
RUN mkdir -p /app/.cache
# Tzdata is installed to allow setting of the timezone via the environment variables
RUN apk add bash tzdata && rm -rf /var/cache/apk/*

COPY motd /etc/motd
COPY .bashrc /app/.bashrc
ENV ENV=/app/.bashrc

RUN chown -R smarthome:smarthome /app
WORKDIR /app/

HEALTHCHECK --interval=300s --timeout=5s --start-period=5s CMD curl --fail http://localhost:80/health || exit 1

EXPOSE 80/tcp

USER smarthome:smarthome

ENV SMARTHOME_LOG_LEVEL=INFO
ENV SMARTHOME_ENV_PRODUCTION=TRUE

ENTRYPOINT ["/app/smarthome"]
