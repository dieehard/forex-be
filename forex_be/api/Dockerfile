FROM golang

RUN go get "github.com/dieehard/forex-be"
WORKDIR "/go/src/github.com/dieehard/forex-be/"

ADD . ./
RUN CGO_ENABLED=0 go build -o api_main

FROM scratch
COPY --from=0 "/go/src/github.com/dieehard/forex-be/api_main" /

ENTRYPOINT ["/api_main"]

EXPOSE 8080

