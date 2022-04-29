FROM alpine:3.14
RUN mkdir -p kubeedge
COPY ./bin/Template kubeedge/bin/
COPY ./res kubeedge/res/
WORKDIR kubeedge/bin
ENTRYPOINT ["./Template"]