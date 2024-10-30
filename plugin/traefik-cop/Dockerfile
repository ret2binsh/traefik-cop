FROM golang:alpine as dev
WORKDIR /opt/app
COPY go.mod go.sum main.go /opt/app
ENV GOARCH=amd64
RUN go mod tidy
RUN go build -o traefik-cop main.go

FROM alpine:latest
WORKDIR /app
COPY --from=dev /opt/app/traefik-cop /app/
EXPOSE 7000
RUN chmod +x /app/traefik-cop
ENTRYPOINT ["./traefik-cop"]
