FROM alpine:3.18

RUN apk add --no-cache \
            bash \
            curl

COPY --from=ghcr.io/equinix-labs/otel-cli:v0.3.0 /otel-cli /usr/bin/otel-cli

COPY simulate-traffic.sh /simulate-traffic.sh

ENTRYPOINT ["/simulate-traffic.sh"]
