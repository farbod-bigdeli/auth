FROM golang:1.17 as builder

ENV GO111MODULE=on
WORKDIR /app
COPY go.mod .
COPY go.sum .

RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/main

FROM alpine 
RUN mkdir /app
RUN cd /app && mkdir storage
RUN mkdir storage
COPY --from=builder /app/main /app/
COPY .env /app/
COPY config_server.json /app/


EXPOSE 8080
# CMD ["go","run","."]
ENTRYPOINT [ "/app/main" ]

# EXPOSE 8080
# CMD ["go","run","."]