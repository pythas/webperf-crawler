FROM golang:1.22-alpine

WORKDIR /app

COPY . .

RUN go build -o main .

COPY --chmod=777 ./run.sh /etc/periodic/daily/run

CMD [ "/usr/sbin/crond", "-f", "-d8" ]