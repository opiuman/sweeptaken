
FROM alpine

WORKDIR /go
COPY sweeptaken /go/sweeptaken

ENTRYPOINT ["./sweeptaken"]