FROM golang:1.27-alpine AS build
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /server ./cmd/server

FROM alpine:3.22
RUN adduser -D app
USER app
COPY --from=build /server /server
EXPOSE 8080
HEALTHCHECK CMD wget -qO- http://localhost:8080/api/health || exit 1
CMD ["/server"]
