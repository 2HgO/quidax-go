FROM golang:1.22.4

RUN mkdir -p /app

WORKDIR /app

COPY go.mod .

COPY go.sum .

RUN go mod download

COPY . .

CMD [ "go", "run", "." ]
