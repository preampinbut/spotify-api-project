FROM golang:1.24.2 AS base
WORKDIR /app

FROM base AS build
COPY . .
RUN go run github.com/steebchen/prisma-client-go generate
RUN go build -v -o nodaemon ./cmd/nodaemon/

FROM base
COPY --from=build /app/nodaemon /usr/local/bin/app
CMD [ "app" ]
