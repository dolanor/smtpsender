# syntax=docker/dockerfile:1
FROM golang:1.21 AS builder

WORKDIR /app

ENV GOMODCACHE /root/.cache/gocache
RUN --mount=target=. --mount=type=cache,target=/root/.cache \
    go install

FROM gcr.io/distroless/base-debian12:latest

COPY --from=builder /go/bin/smtpsender /bin/smtpsender

EXPOSE 12345

CMD [ "/bin/smtpsender" ]
