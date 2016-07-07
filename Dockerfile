FROM scratch
MAINTAINER Gemnasium "support@gemnasium.com"
ADD ca-certificates.crt /etc/ssl/certs/
ADD toolbelt /toolbelt
ENTRYPOINT ["/toolbelt"]
