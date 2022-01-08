FROM alpine:latest
RUN apk add --update ca-certificates

ADD k8s-leader-election /usr/bin/k8s-leader-election

ENTRYPOINT ["/usr/bin/k8s-leader-election"]
