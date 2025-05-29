FROM golang

WORKDIR /app

COPY ./config /app/config 
COPY ./cmd /app/cmd 
COPY ./internal /app/internal
COPY ./go.mod /app

ENV GOPRIVATE=github.com/IlianBuh
ENV CONFIG_PATH="/app/config/config.yml"

EXPOSE 20202

RUN go env -w GOPRIVATE="github.com/IlianBuh"
RUN go mod tidy
RUN go install /app/cmd/sso/main.go

CMD main