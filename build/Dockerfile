# golang docker build
FROM golang:1.21-alpine3.18 as build

WORKDIR /app
COPY . .
RUN apk update && apk add --no-cache ca-certificates

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags '-s' -o main main.go

FROM golang:1.21-alpine3.18 as build
WORKDIR /app

COPY --from=build /app/config/app.yaml ./config/app.yaml
COPY --from=build /app/abi ./abi
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /app/main ./
RUN ls -lah

# Command to run the executable
ENTRYPOINT ["./main"]