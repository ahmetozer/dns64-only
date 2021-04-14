FROM golang
WORKDIR /app
COPY . .
RUN go get -v . 
RUN CGO_ENABLED=0 go build -o /bin/dns64-only

FROM scratch
LABEL org.opencontainers.image.source="https://github.com/ahmetozer/dns64-only"
COPY --from=0 /bin/dns64-only /bin/dns64-only
ENTRYPOINT [ "/bin/dns64-only" ]