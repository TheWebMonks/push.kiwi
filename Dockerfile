#
# Development image
#
#FROM debian:jessie
FROM golang:1.7.3-wheezy

EXPOSE 8080 443

# Packaged dependencies
RUN apt-get update && \
    apt-get install -y \
	curl \
	git \
	tar && \
	rm -rf /var/lib/apt/lists/*

# Package manager:
#   https://github.com/kardianos/govendor
# Include static files in the binary:
#   https://github.com/jteeuwen/go-bindata
RUN go get -u github.com/kardianos/govendor && \
    go get -u github.com/jteeuwen/go-bindata/...

# Set workdir
WORKDIR /go/src/github.com/lukin0110/push.kiwi

# Add the entrypoint.sh
COPY deployment/docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh
RUN chmod ugo+x /usr/local/bin/docker-entrypoint.sh
ENTRYPOINT ["docker-entrypoint.sh"]

# Copy the source
COPY . /go/src/github.com/lukin0110/push.kiwi

# Run bash shell by default
CMD ["bash"]
