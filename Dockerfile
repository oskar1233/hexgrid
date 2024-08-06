FROM golang:1.22.5-alpine

RUN go install github.com/a-h/templ/cmd/templ@latest &&\
    apk add npm nodejs

WORKDIR /app
COPY go.mod go.sum .
RUN go mod download

COPY package.json package-lock.json .

COPY *.go *.templ input.css tailwind.config.js .
COPY static/ static/
RUN templ generate
RUN npx tailwindcss -i ./input.css -o ./static/style.css
RUN CGO_ENABLED=0 GOOS=linux go build -o /hexgrid

EXPOSE 8080

CMD ["/hexgrid"]
