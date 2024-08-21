FROM golang:1.22-alpine
RUN go version
ENV GOPATH=/
COPY ./ ./
RUN go mod download
RUN go build -o auth-service ./cmd/sso/main.go
CMD ["./auth-service"]
EXPOSE 8080