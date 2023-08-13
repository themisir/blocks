##
## BUILD
##
FROM golang:1.21-alpine AS build
RUN apk add --update --no-cache gcc musl-dev

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

WORKDIR /

COPY --from=build /build/app /app

ENV ADDRESS=":80"
EXPOSE 80

CMD ["/app"]
