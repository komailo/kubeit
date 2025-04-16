FROM golang:1.24 AS build

WORKDIR /workdir

RUN sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d

ENV PATH="/workdir/bin:${PATH}"

COPY go.mod go.sum /workdir/
RUN go mod download

COPY . /workdir/
RUN task build

FROM scratch

WORKDIR /app

COPY --from=build /workdir/build/reflow /app/reflow

ENTRYPOINT ["/app/reflow"]
