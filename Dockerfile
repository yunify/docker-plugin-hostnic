FROM alpine:3.4
MAINTAINER jolestar <jolestar@gmail.com>

COPY bin/alpine/docker-plugin-hostinic /usr/bin/

VOLUME /run/docker/plugins

CMD ["/usr/bin/docker-plugin-hostinic"]
