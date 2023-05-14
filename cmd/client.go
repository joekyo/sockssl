package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"sockssl"
)

var (
	socksAddr  string
	serverAddr string

	rootCaPem string
)

const (
	defaultSocksInterface = "127.0.0.1"
	defaultSocksPort      = "1080"

	defaultServerPort = "2080"

	defaultRootCaPem = "root-ca.pem"
)

func init() {
	socksInterface := flag.String("i", defaultSocksInterface, "listen interface")
	socksPort := flag.String("p", defaultSocksPort, "listen port")
	flag.StringVar(&rootCaPem, "ca", defaultRootCaPem, "root CA")
	flag.Parse()
	socksAddr = net.JoinHostPort(*socksInterface, *socksPort)

	log.SetFlags(log.Ltime)
	if flag.NArg() != 1 {
		log.Fatalln("Missing command line argument `host[:port]`")
	}
	serverAddr = flag.Args()[0]
	if !strings.Contains(serverAddr, ":") {
		serverAddr = net.JoinHostPort(serverAddr, defaultServerPort)
	}
}

func main() {
	servername, _, _ := net.SplitHostPort(serverAddr)

	pem, err := os.ReadFile(rootCaPem)
	if err != nil {
		log.Fatal("Failed to read root CA")
	}
	certPool := x509.NewCertPool()
	if ok := certPool.AppendCertsFromPEM(pem); !ok {
		log.Fatal("Failed to add root CA to cert pool")
	}

	config := &tls.Config{
		ServerName: servername,
		RootCAs: certPool,
		ClientSessionCache: tls.NewLRUClientSessionCache(32),
	}

	ln, err := net.Listen("tcp", socksAddr)
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("SockSSL client serving on %s", socksAddr)
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("Accept connection failed, %v\n", err)
			continue
		}
		go agent(conn, config)
	}
}

func agent(c1 net.Conn, config *tls.Config) {
	defer c1.Close()
	peer := c1.RemoteAddr()

	target, raw, err := sockssl.Handshake(c1)
	if err != nil {
		log.Printf("%s ! Handshake failed\n", peer)
		return
	}
	log.Printf("%s > %s\n", peer, target)

	c2, err := net.Dial("tcp", serverAddr)
	if err != nil {
		log.Printf("%s ! Could not connect to SockSSL server %s\n", peer, serverAddr)
		return
	}

	c2 = tls.Client(c2, config)
	defer c2.Close()

	if _, err = c2.Write(raw); err != nil {
		log.Printf("%s ! Could not write to SockSSL server %s: %s\n", peer, serverAddr, err)
		return
	}
	log.Printf("%s = %s\n", peer, target)

	start := time.Now()
	sent, recv := sockssl.IOCopyLoop(c1, c2)
	elapsed := time.Since(start).Round(time.Second)
	log.Printf("%s x %s (%dB tx, %dB rx, %s)\n", peer, target, sent, recv, elapsed)
}
