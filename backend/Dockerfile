FROM golang:1.22 AS requirements-stage

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN GOOS=linux GOARCH=amd64 go build -ldflags='-s' -o=./bin/linux_amd64/api ./cmd/api

FROM golang:1.22 AS run-test-stage
RUN go test -race -vet=off ./...

FROM golang:1.22

WORKDIR /

COPY --from=requirements-stage ./app/bin/linux_amd64/api ./bin/linux_amd64/api

ENTRYPOINT ["./bin/linux_amd64/api"]