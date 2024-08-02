FROM golang:1.22 AS builder

ENV GO111MODULE=on

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags musl -a -o app-bin

FROM alpine:3.20

WORKDIR /root/

COPY --from=builder /app/app-bin ./

CMD ["./app-bin"] 