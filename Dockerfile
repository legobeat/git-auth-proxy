FROM docker.io/golang:1.22 as builder
RUN mkdir /build
WORKDIR /build
ARG GOPROXY="https://goproxy.io,https://proxy.golang.org,direct"
ENV GOPROXY=$GOPROXY
ADD go.mod go.sum .
RUN go get ./...
ADD . /build/
RUN make build

FROM gcr.io/distroless/static:nonroot
COPY --from=builder /build/git-auth-proxy /app/
WORKDIR /app
USER nonroot:nonroot
ENTRYPOINT ["./git-auth-proxy"]
