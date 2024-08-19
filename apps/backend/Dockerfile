FROM golang:1.23 AS base
WORKDIR /app

FROM base AS build
COPY . .
RUN go build -v -o nodaemon ./cmd/nodaemon/

FROM base
COPY --from=build /app/nodaemon /usr/local/bin/app
CMD [ "app" ]
