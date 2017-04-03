# Push Kiwi

This is the backend for the [Push command](https://github.com/lukin0110/push), written in go. This is a Work In 
Progress. It stores the uploaded files on the local disk. Future releases will allow cloud storage on 
[AWS S3](https://aws.amazon.com/s3/) and/or [Google Cloud Storage](https://cloud.google.com/storage/).

You can share files via email. The server uses [Mailgun](https://mailgun.com) as mail backend.

## Usage

```bash
# Clean expired files
push.kiwi --clean

# Startup with tls
push.kiwi --tls --tlsCert /ssl/kiwi.crt --tlsKey /ssl/kiwi.key
```

## Command line options

| Parameter | Description |
| --- | --- |
| clean | Will clean all expired files |
| version | Show version |
| tls | Use TLS to server, default: *false* |
| tlsCrt | Path to SSL certificate |
| tlsKey | Path to SSL key file |
| rootUrl | Root url that needs to be used when sending emails |
| mailgunDomain | Mailgun domain to send emails from |
| mailgunKey | Private API key of mailgun |
| mailgunPublicKey | Public API key of mailgun  |

## Dev

Build & serve:
```
$ docker-compose run app
$ go build -o server
$ ./server
```

Install packages
```
$ govendor fetch github.com/kennygrant/sanitize@v1.1
```

Shell access
```
$ docker exec -it pushkiwi_app_1 /bin/bash
```

Upload:
```
$ curl -H "x-email: maarten@webmonks.io" --upload-file ./README.md https://push.kiwi/README.md 
```
