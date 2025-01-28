FROM golang:1.22 as builder

ENV APPPATH /app

WORKDIR ${APPPATH}

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code. Note the slash at the end, as explained in
# https://docs.docker.com/reference/dockerfile/#copy
COPY . ${APPPATH}

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o ./admissioncontroller cmd/main.go

FROM golang:alpine

ENV APPPATH /app

WORKDIR ${APPPATH}

COPY --from=builder ${APPPATH}/admissioncontroller ${APPPATH}/admissioncontroller

EXPOSE 8443

CMD [ "/app/admissioncontroller" ]
