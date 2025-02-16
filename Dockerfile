FROM golang:1.23-alpine

WORKDIR /app

COPY . .

RUN go build -v -o /app/bjorn ./src

CMD ["./bjorn"]