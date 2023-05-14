package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"log"
	"net"
	"os"
	"time"

	"sockssl"
)

const (
	defaultServerKey = "site-key.pem"
	defaultServerCert = "site-cert.pem"
	defaultCaCert = "root-ca.pem"

	defaultInterface = "0.0.0.0"
	defaultPort      = "2080"
)

func main() {
	listenInterface := flag.String("i", defaultInterface, "listen interface")
	listenPort := flag.String("p", defaultPort, "listen port")

	caCert := flag.String("ca", defaultCaCert, "server private key")
	serverKey := flag.String("key", defaultServerKey, "server private key")
	serverCert := flag.String("cert", defaultServerCert, "server certificate")
	flag.Parse()

	pem, err := os.ReadFile(*caCert)
	if err != nil {
		log.Fatal("Failed to load CA certificate")
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(pem)

	cert, err := tls.LoadX509KeyPair(*serverCert, *serverKey)
	if err != nil {
		log.Fatalf("Load key and certificate failed, %v\n", err)
	}
	config := &tls.Config{
		ClientCAs: caCertPool,
		ClientAuth: tls.RequireAndVerifyClientCert,
		Certificates:[]tls.Certificate{cert},
	}

	addr := net.JoinHostPort(*listenInterface, *listenPort)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("SockSSL server serving on %s\n", addr)
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("connect failed:", err)
			continue
		}
		go server(conn, config)
	}
}

func server(conn net.Conn, config *tls.Config) {
	c1 := tls.Server(conn, config)
	defer c1.Close()
	peer := c1.RemoteAddr()

	target, _, err := sockssl.DecodeAddress(c1, false)
	if err != nil {
		log.Printf("Invalid address received from %s\n", peer)
		return
	}
	log.Printf("%s > %s\n", peer, target)

	c2, err := net.Dial("tcp", target)
	if err != nil {
		log.Printf("Could not connect to %s\n", target)
		return
	}
	defer c2.Close()
	log.Printf("%s = %s\n", peer, target)

	start := time.Now()
	sent, recv := sockssl.IOCopyLoop(c1, c2)
	elapsed := time.Since(start).Round(time.Second)
	log.Printf("%s x %s (%dB tx, %dB rx, %s)\n", peer, target, sent, recv, elapsed)
}
