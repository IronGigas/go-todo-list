# --- Step1: build an app ---
# using official go image
FROM golang:1.23-alpine AS builder

# set workdir
WORKDIR /src

# copy files
COPY go.mod go.sum ./
# download the dependencies
RUN go mod download

# copy the rest of the files
COPY . .

# comnpile for Linux
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/main .

# --- Step2: create final image ---
FROM alpine:latest

# set work dir
WORKDIR /app

# copy compiled file
COPY --from=builder /app/main .

# copy frontend
COPY web ./web

# set environment variable
ENV TODO_DBFILE="scheduler.db"

# start the app
CMD ["./main"]