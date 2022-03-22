FROM golang:1.17.6

WORKDIR /usr/src/app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

RUN go build -v -o /usr/local/bin/app .

EXPOSE 8000

CMD [ "app" ]