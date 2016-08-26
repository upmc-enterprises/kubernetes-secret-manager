FROM alpine
MAINTAINER Steve Sloka <slokas@upmc.edu>

ADD kubernetes-secret-manager /kubernetes-secret-manager
RUN mkdir -p /var/lib/vault-manager

ENTRYPOINT ["/kubernetes-secret-manager"]
