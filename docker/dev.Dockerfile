FROM golang:1.18

ARG APP_NAME=avito-parser

WORKDIR /app

RUN apt-get install -y git wget && \
    go install github.com/githubnemo/CompileDaemon@latest && \
    wget -q -O - https://dl-ssl.google.com/linux/linux_signing_key.pub | apt-key add - \
    && echo "deb http://dl.google.com/linux/chrome/deb/ stable main" >> /etc/apt/sources.list.d/google.list && \
    apt-get update && apt-get -y install google-chrome-stable

COPY . .

# Copy only config (easier to setup path for config reader)
COPY configs/dev.yml /opt/dev.yml

# Hot reload (polling)
ENTRYPOINT CompileDaemon -polling -build="go build -o ./bin/main ./cmd/main.go" -command="./bin/main --debug --config=/opt/dev.yml"

