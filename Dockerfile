# FROM golang:1.17

# WORKDIR /usr/src/app

# # pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
# COPY go.mod go.sum ./
# RUN go mod download && go mod verify

# COPY . .
# RUN go build -v -o /usr/local/bin/app .

# # EXPOSE 8081

# CMD ["app"]

# syntax=docker/dockerfile:1

FROM golang:1.17.6

WORKDIR /usr/src/app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

RUN go build -v -o /usr/local/bin/app .

EXPOSE 8000

CMD [ "app" ]