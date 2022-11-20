# Build stage
FROM golang:1.19 as build-env
WORKDIR /src/kronos
ADD . /cmd/kronos /src/kronos/
ENV CGO_ENABLED=0
RUN go mod vendor
RUN go build -o /app

FROM zenika/alpine-chrome:with-chromedriver
ENV finn=nope
WORKDIR /
COPY --from=build-env /app /src/kronos/config.json /
RUN ls
USER root
RUN apk update && apk add bash && apk --no-cache add tzdata

ENTRYPOINT ["/app"]
