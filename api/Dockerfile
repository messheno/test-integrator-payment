FROM golang:1.21.1-alpine AS builder
WORKDIR /apps

ADD ./ .
RUN mkdir -p dist/linux

RUN go mod tidy
RUN go install github.com/swaggo/swag/cmd/swag@latest
RUN go get -u github.com/swaggo/swag
RUN swag init -g main.go

RUN go build -a -tags netgo -ldflags "-s -w" -o ./dist/linux/spay .
RUN cp ./app.yaml ./dist/linux/

FROM alpine:3.17
RUN apk add --no-cache bash
WORKDIR /dist
COPY --from=builder /apps/dist/linux/ .

EXPOSE 3000

CMD ["/dist/spay"]