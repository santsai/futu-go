ARG FILE=Futu_OpenD_9.6.5608_Ubuntu18.04

# for extracting opend
FROM debian:trixie-slim AS extractor

RUN apt-get update && \
    apt-get install -y curl procps

# FILE followed from:
#   https://openapi.futunn.com/futu-api-doc/en/opend/opend-cmd.html
#   https://www.futunn.com/download/fetch-lasted-link?name=opend-ubuntu

ARG FILE
RUN curl --output /tmp/opend.tgz \
    -L https://softwaredownload.futunn.com/$FILE.tar.gz

RUN tar -xvf /tmp/opend.tgz -C /tmp


# the real thing
FROM debian:trixie-slim

ARG FILE
WORKDIR /opend
COPY --from=extractor /tmp/$FILE/$FILE/*    /opend
COPY --from=extractor /tmp/$FILE/README.txt /opend

COPY run.sh /opend
RUN chmod u+x /opend/run.sh

EXPOSE 11111/tcp

CMD ["/opend/run.sh"]
