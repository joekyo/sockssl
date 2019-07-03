## SockSSL: secure your SOCKS connection using SSL

## Build from source

Download source code

```shell
git clone https://github.com/joekyo/sockssl.git
cd sockssl
go mod init sockssl
```

on your server

```shell
go build -o sockssl cmd/server.go
./sockssl
```

on your PC or Mac

```shell
go build -o sockssl cmd/client.go
./sockssl example.com
```

## Server side configuration

Before running `sockssl` on server, you need to prepare a valid certificate file and a private key.
You can use tools like `certbot-auto` or `lego` to fetch a free certificate on your server.

When `sockssl` starts, it will try loading the certificate file named `fullchain.pem` and the private key named `key.pem`.
If you have these files with different names, you can use command line flags `-c` and `-k`, e.g.:

```shell
./sockssl -c /path/your_cert -k /path/your_key
```

`sockssl` will listen on default port `2080`, to change it use flag `-p`, e.g.:

```shell
./sockssl -p 8443
```

## Client side configuration

Suppose that your server domain name is `example.com`, to connect to your server simply run

```shell
./sockssl example.com
```

`sockssl` will connect to `example.com` and its default port `2080`. If your server is listening on other port like `8443`, you can run

```shell
./sockssl example.com:8443
```

The client `sockssl` will listening on default interface `127.0.0.1` and port `1080`.
You can use command line flags `-i` and `-p` to change them respectively.
For example, to allow others at same LAN to connect your `sockssl` client, you can run:

```shell
./sockssl -i 0.0.0.0
```