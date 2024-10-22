FROM --platform=$BUILDPLATFORM golang:1.23-alpine AS build-env

# Set working directory for the build
WORKDIR /src

COPY . .

RUN go mod tidy -compat=1.23 && \
    go build /src/cmd/local-sequencer/main.go

# Final image
FROM alpine:3.18.3

WORKDIR /root

# Copy over binaries from the build-env
COPY --from=build-env /src/main /usr/bin/local-sequencer

EXPOSE 7980

CMD ["local-sequencer", "-listen-all"]