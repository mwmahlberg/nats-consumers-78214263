FROM golang:1.22-alpine3.19 as builder
ARG BINARY_NAME="producer"
WORKDIR /tmp/cmds
COPY . .
RUN go build -o ${BINARY_NAME} ./cmds/${BINARY_NAME}

FROM alpine:3.19
ARG BINARY_NAME="producer"
ENV BINARY_NAME=${BINARY_NAME}
COPY --from=builder /tmp/cmds/${BINARY_NAME} /usr/local/bin/${BINARY_NAME}