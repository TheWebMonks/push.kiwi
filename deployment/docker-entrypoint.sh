#!/bin/bash

# Exit immediately if a command exits with a non-zero status.
# http://stackoverflow.com/questions/19622198/what-does-set-e-mean-in-a-bash-script
set -e

# Check if the required environment variables are set
[ -z "$TLS_CERT" ] && echo "ERROR: Need to set TLS_CERT. E.g.: /ssl/push_kiwi.crt" && exit 1;
[ -z "$TLS_KEY" ] && echo "ERROR: Need to set TLS_KEY. E.g.: /ssl/push_kiwi.key" && exit 1;
[ -z "$MAILGUN_DOMAIN" ] && echo "ERROR: Need to set MAILGUN_DOMAIN" && exit 1;
[ -z "$MAILGUN_KEY" ] && echo "ERROR: Need to set MAILGUN_KEY" && exit 1;
[ -z "$MAILGUN_PUBLIC_KEY" ] && echo "ERROR: Need to set MAILGUN_PUBLIC_KEY" && exit 1;

# Define help message
show_help() {
    echo """
Usage: docker run <imagename> COMMAND
Commands:
bash        : Start a bash shell
serve       : Start server on port 8080
serve_tls   : Start TLS server on port 443
build       : Build the server artifact
help        : Show this message
"""
}

LDFLAGS="-X main.BUILD_DATE=`date -u +%Y-%m-%d.%H\:%M\:%S`"

case "$1" in
    bash)
        /bin/bash "${@:2}"
    ;;
    serve)
        echo "Serve ..."
        go generate
        env GOOS=linux go install -ldflags "$LDFLAGS" -v github.com/lukin0110/push.kiwi/
        push.kiwi \
            --root-url $ROOT_URL \
            --mailgun-domain $MAILGUN_DOMAIN \
            --mailgun-key $MAILGUN_KEY \
            --mailgun-public-key $MAILGUN_PUBLIC_KEY
    ;;
    serve_tls)
        echo "Serve TLS ..."
        go generate
        env GOOS=linux go install -ldflags "$LDFLAGS" -v github.com/lukin0110/push.kiwi/
        push.kiwi \
            --tls \
            --tls-cert $TLS_CERT \
            --tls-key $TLS_KEY \
            --root-url $ROOT_URL \
            --mailgun-domain $MAILGUN_DOMAIN \
            --mailgun-key $MAILGUN_KEY \
            --mailgun-public-key $MAILGUN_PUBLIC_KEY
    ;;
    clean)
        go generate
        go build -o server
        ./server --clean
    ;;
    build)
        go generate
        go install -v github.com/lukin0110/push.kiwi/
    ;;
    release)
        echo 'Linux'
        go generate
        #env GOOS=linux go install -ldflags "$LDFLAGS" -v github.com/lukin0110/push.kiwi/

        # We’re disabling cgo which gives us a static binary. We’re also setting the OS to Linux
        # (in case someone builds this on a Mac or Windows) and the -a flag means to rebuild all the packages we’re
        # using, which means all the imports will be rebuilt with cgo disabled. These settings changed in Go 1.4 but
        # I found a workaround in a GitHub Issue. Now we have a static binary!
        # https://blog.codeship.com/building-minimal-docker-containers-for-go-applications/
        env CGO_ENABLED=0 GOOS=linux go install -a -ldflags "$LDFLAGS" -installsuffix cgo -v github.com/lukin0110/push.kiwi/
    ;;
    mac)
        go generate
        env GOOS=darwin go install -ldflags "$LDFLAGS" -v github.com/lukin0110/push.kiwi/
    ;;
    fetch)
        echo 'Fetching (with govendor)'
        govendor fetch "${@:2}"
    ;;
    *)
        show_help
    ;;
esac
