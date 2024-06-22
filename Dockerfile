FROM golang:1.20 as builder
RUN mkdir /build
ADD . /build/
WORKDIR /build
RUN make build

FROM gcr.io/distroless/static:nonroot
COPY --from=builder /build/git-auth-proxy /app/
WORKDIR /app
USER nonroot:nonroot
ENTRYPOINT ["./git-auth-proxy"]
