FROM ubuntu:20.04
WORKDIR /usr/local/bin
COPY ./res /usr/local/res
COPY ./bin /usr/local/bin
RUN echo "HwHiAiUser:x:1000:1000::/home/HwHiAiUser:" >> /etc/passwd ; \
echo "HwHiAiUser:x:1024:" >> /etc/group ; \
echo "HwHiAiUser:!:17795:0:99999:7:::" >> /etc/shadow ; \
chown -R 1000:1000 /usr/local/bin; \
chown -R 1000:1000 /usr/local/res;
USER 1000
ENTRYPOINT ["./gpio","--v","4"]
