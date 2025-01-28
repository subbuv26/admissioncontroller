package main

import (
	"flag"
	"log"

	"admissioncontroller/pkg/controller"
	"admissioncontroller/pkg/validate"
)

var (
	tlsKeyPath  string
	tlsCertPath string
)

func main() {
	validator, err := validate.New()
	if err != nil {
		log.Fatal(err)
	}
	controllerConfig := controller.Config{
		Port:        8443,
		TLSKeyPath:  tlsKeyPath,
		TLSCertPath: tlsCertPath,
	}
	server := controller.New(controllerConfig, validator)

	log.Fatal(server.Start())
}

func init() {
	flag.StringVar(&tlsKeyPath, "tlsKeyPath", "/etc/certs/tls.key", "Absolute path to the TLS key")
	flag.StringVar(&tlsCertPath, "tlsCertPath", "/etc/certs/tls.crt", "Absolute path to the TLS certificate")

	flag.Parse()
}
