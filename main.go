package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/akamensky/argparse"
)

func main() {
	parser := argparse.NewParser("fcbreak-sub", "HTTP Template server for fcbreak service")
	// Create string flag
	infopath := parser.String("i", "info", &argparse.Options{Required: true, Help: "Service info path from fcbreak"})
	addr := parser.String("l", "listen", &argparse.Options{Required: true, Help: "HTTP Listening Address"})
	tls := parser.Flag("s", "tls", &argparse.Options{Help: "Listen HTTPS instead of HTTP"})
	tlscert := parser.String("", "cert", &argparse.Options{Help: "HTTPS Certificate File"})
	tlskey := parser.String("", "key", &argparse.Options{Help: "HTTPS Private Key File"})
	auth := parser.Flag("a", "auth", &argparse.Options{Help: "Enable Authentication"})
	user := parser.String("u", "user", &argparse.Options{Help: "Authentication Username. Can be empty even with auth, all users can access. Empty user can be useful for template rendering."})
	pass := parser.String("p", "pass", &argparse.Options{Help: "Authentication Password. Can be empty even with auth, all passwords can access. Empty password can be useful for template rendering."})
	// Parse input
	err := parser.Parse(os.Args)
	if err != nil {
		// In case of error print error and print usage
		// This can also be done by passing -h or --help flags
		fmt.Print(parser.Usage(err))
		os.Exit(1)
	}

	username := ""
	password := ""
	if user != nil {
		username = *user
	}
	if pass != nil {
		password = *pass
	}

	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)

	renderer := NewRenderer(*infopath, filepath.Join(exPath, "templates"), *auth, username, password)
	fmt.Println("Server Starting")
	http.HandleFunc("/", renderer.ServeHTTP)

	if tls != nil && *tls {
		if tlscert == nil || tlskey == nil {
			fmt.Print(parser.Usage(errors.New("tls cert or key is missing")))
			os.Exit(1)
		}
		http.ListenAndServeTLS(*addr, *tlscert, *tlskey, nil)
	} else {
		http.ListenAndServe(*addr, nil)
	}
}
