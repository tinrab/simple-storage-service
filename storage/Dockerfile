FROM golang:1.9.2

WORKDIR /go/src/app
COPY . .

RUN go install -v .

EXPOSE 3000

CMD [ "app" ]
