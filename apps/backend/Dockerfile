FROM golang:1.23 AS base
WORKDIR /app

FROM base AS build
COPY . .
RUN go build -v -o daemon ./cmd/daemon/

FROM base
COPY --from=build /app/daemon /usr/local/bin/app
CMD [ "app" ]
