// Copyright (c) 2018, Sylabs Inc. All rights reserved.
// This software is licensed under a 3-clause BSD license. Please consult the
// LICENSE file distributed with the sources of this project regarding your
// rights to use or distribute this software.

package main

import (
	"context"
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sylabs/cri/pkg/image"
	"github.com/sylabs/cri/pkg/runtime"
	"google.golang.org/grpc"
	"k8s.io/kubernetes/pkg/kubectl/util/logs"
	k8s "k8s.io/kubernetes/pkg/kubelet/apis/cri/runtime/v1alpha2"
)

type flags struct {
	socket   string
	storeDir string
}

func readFlags() flags {
	var f flags
	flag.StringVar(&f.socket, "sock", "/var/run/singularity.sock", "unix socket to serve cri services")
	flag.StringVar(&f.storeDir, "store", "/var/lib/singularity", "directory to store all pulled images")
	flag.Parse()
	return f
}

func logGRPC(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()
	resp, err := handler(ctx, req)
	log.Printf("Method:%s\n\tRequest: %v\n\tResponse: %v\n\tError: %v\n\tDuration:%s\n",
		info.FullMethod, req, resp, err, time.Since(start))
	return resp, err
}

func main() {
	f := readFlags()
	logs.InitLogs()
	defer logs.FlushLogs()

	exitCh := make(chan os.Signal, 1)
	signal.Notify(exitCh, syscall.SIGINT, syscall.SIGTERM)

	lis, err := net.Listen("unix", f.socket)
	if err != nil {
		log.Fatalf("Could not start CRI listener: %v ", err)
	}
	defer lis.Close()

	syRuntime, err := runtime.NewSingularityRuntime()
	if err != nil {
		log.Printf("Could not create Singularity runtime service: %v", err)
		return
	}
	syImage, err := image.NewSingularityRegistry(f.storeDir)
	if err != nil {
		log.Printf("Could not create Singularity image service: %v", err)
		return
	}

	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(logGRPC))
	k8s.RegisterRuntimeServiceServer(grpcServer, syRuntime)
	k8s.RegisterImageServiceServer(grpcServer, syImage)

	log.Printf("Singularity CRI server started on %v", lis.Addr())
	go grpcServer.Serve(lis)

	<-exitCh

	log.Println("Singularity CRI service exiting...")
}
