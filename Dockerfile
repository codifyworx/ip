FROM golang:1.22-alpine AS build

WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/codify-ip .

FROM alpine:3.20

RUN adduser -D -H -u 10001 appuser
USER appuser

COPY --from=build /out/codify-ip /usr/local/bin/codify-ip

EXPOSE 8080
ENV ADDR=:8080
ENV BASE_PATH=/ip

ENTRYPOINT ["/usr/local/bin/codify-ip"]
