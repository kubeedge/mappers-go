FROM ubuntu:20.04
WORKDIR /usr/local/bin
COPY ./res /usr/local/res
COPY ./bin/idmvs /usr/local/bin/idmvs
ENTRYPOINT ["./idmvs", "--v", "4"]