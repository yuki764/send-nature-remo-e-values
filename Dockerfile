ARG GO_VERSION=1.20.2
FROM golang:${GO_VERSION}-bullseye as build

WORKDIR /go/src/app
COPY . .
RUN go mod download
ENV CGO_ENABLED=0
RUN go build ./cmd/send-nature-remo-e-values

FROM gcr.io/distroless/static-debian11
COPY --from=build /go/src/app/send-nature-remo-e-values /
CMD ["/send-nature-remo-e-values"]
