FROM arm64v8/golang:1.22.6-alpine AS builder

RUN apk --no-cache add bash git make gcc gettext musl-dev libwebp-dev

ARG goprivate
ARG machine
ARG login
ARG password

ENV GOPRIVATE=$goprivate

RUN echo "machine $machine" > ~/.netrc && \
    echo "    login $login" >> ~/.netrc && \
    echo "    password $password" >> ~/.netrc

WORKDIR /usr/local/src/

# cache dependencies
COPY ["app/go.mod", "app/go.sum", "./"]
RUN go mod download

# build application
COPY app ./
RUN go build -o ./bin/app app/cmd/software/main.go

FROM alpine

RUN apk update && \
    apk upgrade -U && \
    apk add libwebp-dev  && \
    rm -rf /var/cache/*

COPY --from=builder /usr/local/src/bin/app /
COPY configs /configs

CMD ["/app"]