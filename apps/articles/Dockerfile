FROM golang:1.18-alpine

WORKDIR /articles-microservice

COPY . .

RUN go mod download

RUN go build -o articles-microservice main.go

CMD [ "./articles-microservice" ]