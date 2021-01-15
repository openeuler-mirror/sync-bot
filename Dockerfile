# Build stage
FROM golang:1.15.5 AS build

ADD . /go-build

WORKDIR /go-build

ENV GOPROXY=https://goproxy.cn,direct

RUN go build -o /sync-bot


# Final stage
FROM centos:8

RUN dnf -y install git

RUN git config --global user.name openeuler-sync-bot

RUN git config --global user.email openeuler.syncbot@gmail.com

EXPOSE 8765

WORKDIR /

COPY --from=build /sync-bot /

CMD ["/sync-bot"]
