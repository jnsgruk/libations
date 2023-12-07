package main

//go:generate bash -c "hugo -s ./webui --minify"

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"

	"tailscale.com/tsnet"
)

var (
	//go:embed webui/public
	site embed.FS

	addr     = flag.String("addr", ":443", "address to listen on")
	hostname = flag.String("hostname", "libations", "hostname to use on the tailnet")
)

func redirectToTLS(w http.ResponseWriter, r *http.Request) {
	newURL := fmt.Sprintf("https://%s%s", r.Host, r.RequestURI)
	http.Redirect(w, r, newURL, http.StatusMovedPermanently)
}

func main() {
	flag.Parse()
	s := &tsnet.Server{Hostname: *hostname}
	defer s.Close()

	// Start a standard HTTP server in the background to redirect HTTP -> HTTPS
	go func() {
		httpLn, err := s.Listen("tcp", ":80")
		if err != nil {
			log.Fatalln(err)
		}

		err = http.Serve(httpLn, http.HandlerFunc(redirectToTLS))
		if err != nil {
			log.Fatalln(err)
		}
	}()

	tlsLn, err := s.ListenTLS("tcp", *addr)
	if err != nil {
		log.Fatalln(err)
	}
	defer tlsLn.Close()

	// Create an fs.FS from the embedded filesystem
	fSys, err := fs.Sub(site, "webui/public")
	if err != nil {
		log.Fatal(err)
	}

	// Serve an HTTP file server over TLS using our embedded filesystem
	err = http.Serve(tlsLn, http.FileServer(http.FS(fSys)))
	if err != nil {
		log.Fatalln(err)
	}
}
