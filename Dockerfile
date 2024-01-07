FROM golang:1.21 AS build

WORKDIR /app

COPY go.* ./

RUN go mod download

COPY ./main.go ./
COPY ./driverConfig.go ./
COPY ./persistence.go ./

RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ./rinha .

FROM alpine:3.14.10

EXPOSE 8080

COPY --from=build /app/rinha .

CMD ["/rinha"]
