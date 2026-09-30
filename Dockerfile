FROM golang:1.26-alpine as BUILDER

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o api ./cmd/api
RUN CGO_ENABLED=0 go build -o worker ./cmd/worker

FROM golang:1.26-alpine as RUNNER

WORKDIR /app

RUN apk --no-cache add ca-certificates

COPY --from=BUILDER ./app/api .
COPY --from=BUILDER ./app/worker .

EXPOSE 8080
CMD ["./api"]

