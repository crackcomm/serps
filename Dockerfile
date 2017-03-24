FROM alpine:latest
MAINTAINER Łukasz Kurowski <crackcomm@gmail.com>

# Copy application
COPY ./dist/serps /serps

#
# Configuration environment variables
#
ENV NSQ_ADDR nsq:4150
ENV NSQLOOKUP_ADDR nsqlookup:4161

ENTRYPOINT ["/serps"]
