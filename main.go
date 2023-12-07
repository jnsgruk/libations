package main

//go:generate bash -c "hugo -s ./webui --minify"

import (
	"embed"
	"flag"
	"fmt"
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

func routeHandler(w http.ResponseWriter, r *http.Request) {
	path := fmt.Sprintf("webui/public%s", r.URL.Path)
	if path == "webui/public/" {
		path = "webui/public/index.html"
	}

	data, err := site.ReadFile(path)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "File not found: %s", path)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func main() {
	flag.Parse()

	s := new(tsnet.Server)
	s.Hostname = *hostname
	defer s.Close()

	ln, err := s.ListenTLS("tcp", *addr)
	if err != nil {
		log.Fatalln(err)
	}
	defer ln.Close()

	err = http.Serve(ln, http.HandlerFunc(routeHandler))
	if err != nil {
		log.Fatalln(err)
	}
}
