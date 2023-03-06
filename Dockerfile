FROM golang:1.19-alpine
WORKDIR /
COPY ./ .
RUN apk add --no-cache make
RUN go mod download

RUN make build/api

EXPOSE 4000

CMD ["make", "start"]
