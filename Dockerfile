FROM golang:1.15 as build
WORKDIR /opt/src
COPY . .
RUN groupadd -g 1000 appuser &&\
    useradd -m -u 1000 -g appuser appuser
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o /opt/action-github-deployment

FROM scratch
LABEL "repository"="https://github.com/ReasonSoftware/action-github-deployment"
LABEL "version"="1.0.0"
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /etc/passwd /etc/passwd
COPY LICENSE.md /LICENSE.md
COPY --from=build --chown=1000:0 /opt/action-github-deployment /app
ENTRYPOINT [ "/app" ]