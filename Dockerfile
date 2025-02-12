FROM golang:1.24-alpine3.20 AS build

WORKDIR /app
COPY go.* ./
RUN go mod download

COPY . .
RUN go get -d -v && GOOS=linux GOARCH=amd64 go build -v

FROM alpine:3.21 AS run

RUN apk add --no-cache tzdata
ENV TZ=Europe/Berlin

ENV UID=65532
ENV GID=65532
ENV USER=nonroot
ENV GROUP=nonroot

RUN addgroup -g $GID $GROUP && \
    adduser --shell /sbin/nologin --disabled-password \
    --no-create-home --uid $UID --ingroup $GROUP $USER

WORKDIR /app/
COPY --from=build /app/ths-proxy /app/app.yml ./
USER $USER

ENV GIN_MODE=release
EXPOSE 8080

ENTRYPOINT ["/app/ths-proxy"]
