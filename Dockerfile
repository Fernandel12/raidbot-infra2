FROM            golang:1.23.6-alpine AS gobuilder
RUN             apk --no-cache --update add npm make gcc g++ musl-dev openssl-dev git perl-utils curl
WORKDIR         /go/src/rslbot
ENV             GO111MODULE=on GOPROXY=https://proxy.golang.org,direct
COPY            go.mod go.sum ./
RUN             go mod download
COPY            . .
WORKDIR         /go/src/rslbot/go
RUN             make install

# runtime
FROM            alpine:3.20
RUN             apk --no-cache --update add openssl wget bash
RUN             wget https://raw.githubusercontent.com/vishnubob/wait-for-it/master/wait-for-it.sh && chmod +x wait-for-it.sh
COPY            --from=gobuilder /go/bin/rslbot /bin/rslbot
ENTRYPOINT      ["/bin/rslbot"]
CMD             ["api"]
EXPOSE          8000 9111
