FROM golang:1.8

ADD deploy.sh /go/

CMD ["./deploy.sh"]
