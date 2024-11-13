FROM golang:1.22.4

# RUN apk add build-base
# RUN apk --no-cache add openssl

RUN mkdir -p /app

WORKDIR /app

COPY go.mod .

COPY go.sum .

RUN go mod download

COPY . .

RUN GIT_TERMINAL_PROMPT=1 \
    # CGO_CFLAGS='-O2 -g -w' \
    CGO_ENABLED=1 \
    go build -v -o quidax-go

CMD ["./quidax-go"]
