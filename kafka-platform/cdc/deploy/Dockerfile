FROM alpine:3.19

RUN apk add --no-cache bash curl jq

COPY scripts/wait-and-deploy.sh /usr/local/bin/wait-and-deploy.sh
RUN chmod +x /usr/local/bin/wait-and-deploy.sh

ENTRYPOINT ["/usr/local/bin/wait-and-deploy.sh"]