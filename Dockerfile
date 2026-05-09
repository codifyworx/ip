FROM golang:alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/codify-ip .

FROM scratch

COPY --from=build /out/codify-ip /usr/local/bin/codify-ip

USER 10001
EXPOSE 8080
ENV ADDR=:8080
ENV BASE_PATH=/ip

ENTRYPOINT ["/usr/local/bin/codify-ip"]
