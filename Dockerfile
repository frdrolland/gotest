#FROM golang:onbuild
FROM golang

ENV http_proxy http://10.79.247.87:3128/
ENV https_proxy http://10.79.247.87:3128/

EXPOSE 8080

RUN mkdir /testfro
COPY gotest /testfro/

ENTRYPOINT ["gotest"]
