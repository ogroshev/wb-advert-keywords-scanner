FROM golang:1.19

WORKDIR /usr/src/app

COPY go.mod go.sum ./
ARG GITLAB_TOKEN
RUN git config --global url."https://oauth2:${GITLAB_TOKEN}@gitlab.com/wb-dynamics".insteadOf "https://gitlab.com/wb-dynamics"
RUN go env -w GOPRIVATE="gitlab.com/wb-dynamics/wb-go-libs"
RUN go mod download && go mod verify

COPY . .
RUN cd cmd; go build -v -o /usr/local/bin/app .

CMD ["app"]