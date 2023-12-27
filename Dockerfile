FROM golang:1.20 AS build

# Create a group and user
RUN groupadd -r appuser && useradd -r -u 10000 -g appuser appuser

WORKDIR /app

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /discordcalorietracker

# Deploy the application binary into a lean image
FROM scratch AS release

WORKDIR /

COPY --from=build /discordcalorietracker /discordcalorietracker

# copy ca certs
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
 
# copy users/groups from builder
COPY --from=build /etc/passwd /etc/passwd
COPY --from=build /etc/group /etc/group

USER appuser:appuser

ENTRYPOINT ["/discordcalorietracker"]
CMD ["-token"]