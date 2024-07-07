##
## BUILD
##
FROM golang:1.21-bookworm AS build
RUN aptget install -y gcc

WORKDIR /build

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 go build -o app .

##
## DEPLOY
##
FROM alpine:latest AS deploy
RUN apk add --update --no-cache tzdata

WORKDIR /

COPY --from=build /build/app /app

ENV ADDRESS=":80"
EXPOSE 80

CMD ["/app"]
