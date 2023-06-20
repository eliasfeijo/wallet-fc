FROM golang:1.20

WORKDIR /app/

RUN apt-get update && apt-get install -y wget librdkafka-dev

ENV DOCKERIZE_VERSION v0.7.0
RUN wget https://github.com/jwilder/dockerize/releases/download/$DOCKERIZE_VERSION/dockerize-linux-amd64-$DOCKERIZE_VERSION.tar.gz \
    && tar -C /usr/local/bin -xzvf dockerize-linux-amd64-$DOCKERIZE_VERSION.tar.gz \
    && rm dockerize-linux-amd64-$DOCKERIZE_VERSION.tar.gz

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o balances-ms ./cmd/wallet_consumer

CMD ["./balances-ms"]