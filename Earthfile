groundcover-deps:
    FROM golang:1.18-alpine3.15
    RUN apk add build-base
    WORKDIR /builder
    COPY go.mod .
    COPY go.sum .
    RUN go mod download
    # Output these back in case go mod download changes them.
    SAVE ARTIFACT go.mod AS LOCAL go.mod
    SAVE ARTIFACT go.sum AS LOCAL go.sum


pkg-base:
    FROM +groundcover-deps
    COPY --dir pkg .


build-cli:
    FROM +pkg-base
    COPY --dir cmd main.go .
    ARG EARTHLY_GIT_HASH
    ARG IMAGE_TAG=$EARTHLY_GIT_HASH
    RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-X 'groundcover.com/cmd.BinaryVersion=${IMAGE_TAG}'" -o /bin/groundcover ./main.go
    SAVE ARTIFACT /bin/groundcover AS LOCAL ./artifacts/groundcover

build-cli-image:
    FROM alpine/helm:3.9.0
    COPY +build-cli/groundcover ./
    ENTRYPOINT ["./groundcover"]
    ARG EARTHLY_GIT_HASH
    ARG IMAGE_TAG=$EARTHLY_GIT_HASH
    ARG REG=125608480246.dkr.ecr.eu-west-3.amazonaws.com
    SAVE IMAGE --push $REG/groundcover-cli:$IMAGE_TAG
