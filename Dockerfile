# Build stage
FROM golang:1.15.5 AS build

ADD . /go-build

WORKDIR /go-build

ENV GOPROXY=https://goproxy.cn,direct

RUN go build -o /sync-bot


# Final stage
FROM openeuler/openeuler:22.03-lts

RUN dnf -y install git

RUN git config --global user.name openeuler-sync-bot

RUN git config --global user.email openeuler.syncbot@gmail.com

EXPOSE 8765

WORKDIR /

COPY --from=build /sync-bot /
COPY --from=build /drop_branches.config /

# ADD secret.conf /
# ADD token.conf /

CMD ["/sync-bot"]
