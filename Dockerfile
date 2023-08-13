##
## BUILD
##
FROM golang:1.21 AS build

WORKDIR /build

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 go build -o /blocks .

##
## DEPLOY
##
FROM gcr.io/distroless/static:latest AS deploy

WORKDIR /

COPY --from=build /blocks /blocks

ENV ADDRESS=":80"
USER nonroot:nonroot
EXPOSE 80

CMD ["/blocks"]
