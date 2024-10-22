package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/rollkit/go-sequencing/proxy/grpc"
	"github.com/rollkit/go-sequencing/test"
)

const (
	defaultHost     = "localhost"
	defaultPort     = "50051"
	defaultRollupId = "rollup-id"
)

func main() {
	var (
		host      string
		port      string
		rollupId  string
		listenAll bool
	)
	flag.StringVar(&port, "port", defaultPort, "listening port")
	flag.StringVar(&host, "host", defaultHost, "listening address")
	flag.StringVar(&rollupId, "rollup-id", defaultRollupId, "rollup id")
	flag.BoolVar(&listenAll, "listen-all", false, "listen on all network interfaces (0.0.0.0) instead of just localhost")
	flag.Parse()

	if listenAll {
		host = "0.0.0.0"
	}

	d := test.NewDummySequencer([]byte(rollupId))
	srv := grpc.NewServer(d, d, d)
	log.Printf("Listening on: %s:%s", host, port)

	listenAddress := fmt.Sprintf("%s:%s", host, port)
	lis, err := net.Listen("tcp", listenAddress)
	if err != nil {
		log.Fatal("error while serving:", err)
	}
	go func() {
		_ = srv.Serve(lis)
	}()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGINT)
	<-interrupt
	fmt.Println("\nCtrl+C pressed. Exiting...")
	os.Exit(0)
}
