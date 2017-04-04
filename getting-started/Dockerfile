FROM alpine:3.3

RUN apk update
RUN apk add bash iperf

ADD squirrel-worker /bin/

ENTRYPOINT ["/bin/squirrel-worker"]
