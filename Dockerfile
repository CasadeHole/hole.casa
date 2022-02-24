FROM golang:1.17 AS builder

WORKDIR /app

COPY ./go.mod ./go.sum ./

RUN go mod download

COPY ./ ./

RUN go build -o web .

FROM gcr.io/distroless/base:nonroot
COPY --from=builder /app/web /web
ENTRYPOINT [ "/web" ]