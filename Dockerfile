FROM alpine:latest

COPY bin/operator /usr/local/bin/operator

ENTRYPOINT ["/usr/local/bin/operator"]
