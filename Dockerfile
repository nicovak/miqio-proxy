##############################################
# STAGE 1: Go release
##############################################
ARG ALPINE_VERSION=3.23
ARG GOLANG_VERSION=1.26.2

FROM golang:${GOLANG_VERSION}-alpine${ALPINE_VERSION} as builder

ARG GORELEASER_VERSION=2.15.0

ENV APP_USER=miqio-proxy

RUN apk add --update git build-base wget
RUN wget https://github.com/goreleaser/goreleaser/releases/download/v${GORELEASER_VERSION}/goreleaser_${GORELEASER_VERSION}_x86_64.apk
RUN apk add --allow-untrusted goreleaser_${GORELEASER_VERSION}_x86_64.apk

RUN --mount=type=secret,id=github source /run/secrets/github \
    && git config --global url."https://${GITHUB_PAT}:@github.com/".insteadOf "https://github.com/"

WORKDIR /go/src/github.com/nicovak/${APP_USER}

COPY . .

RUN goreleaser build --snapshot

##############################################
# STAGE 2: Final container
##############################################
ARG ALPINE_VERSION=3.23

FROM alpine:${ALPINE_VERSION}
ARG APP=/app

ENV TZ=Europe/Paris
ENV APP_USER=miqio-proxy

RUN addgroup -S ${APP_USER} \
    && adduser -S ${APP_USER} -G ${APP_USER} \
    && mkdir -p ${APP}

RUN chown -R ${APP_USER}:${APP_USER} ${APP}

WORKDIR /app/$APP_USER

COPY --from=builder "/go/src/github.com/nicovak/${APP_USER}/dist/${APP_USER}_linux_amd64_v1/${APP_USER}" .
COPY --from=builder "/go/src/github.com/nicovak/${APP_USER}/db" "./db"

RUN touch .env

USER $APP_USER

CMD ./miqio-proxy server
