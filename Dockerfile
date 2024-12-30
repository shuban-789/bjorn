FROM golang:1.23-alpine

WORKDIR /app

COPY . .

RUN --mount=type=secret,id=dsc_bot_token go build -v -o /app/bjorn ./src

CMD ["./bjorn"]