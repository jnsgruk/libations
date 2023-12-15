package main

//go:generate bash -c "hugo -s ./webui --minify"

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"

	"tailscale.com/tsnet"
)

var (
	//go:embed webui/public
	site embed.FS

	addr      = flag.String("addr", ":443", "address to listen on")
	hostname  = flag.String("hostname", "libations", "hostname to use on the tailnet")
	tsnetLogs = flag.Bool("tsnet-logs", true, "include tsnet logs in application logs")
)

func redirectToTLS(w http.ResponseWriter, r *http.Request) {
	newURL := fmt.Sprintf("https://%s%s", r.Host, r.RequestURI)
	http.Redirect(w, r, newURL, http.StatusMovedPermanently)
}

func main() {
	flag.Parse()
	log := slog.Default().With(slog.String("source", "libations"))

	s := &tsnet.Server{
		Hostname: *hostname,
		Logf: func(msg string, args ...any) {
			l := slog.Default().With(
				slog.String("source", "tsnet"),
				slog.String("hostname", *hostname),
			)
			l.Info(fmt.Sprintf(msg, args...))
		},
	}
	defer s.Close()

	if !*tsnetLogs {
		log.Warn("tsnet logs are disabled, interactive auth link will not be shown")
		s.Logf = func(string, ...any) {}
	}

	// Start a standard HTTP server in the background to redirect HTTP -> HTTPS
	go func() {
		log.Info(fmt.Sprintf("starting HTTP listener on tsnet %s:80", *hostname))
		httpLn, err := s.Listen("tcp", ":80")
		if err != nil {
			log.Error(err.Error())
			os.Exit(1)
		}

		err = http.Serve(httpLn, http.HandlerFunc(redirectToTLS))
		if err != nil {
			log.Error(err.Error())
			os.Exit(1)
		}
	}()

	log.Info(fmt.Sprintf("starting HTTPS listener on tsnet %s:443", *hostname))
	tlsLn, err := s.ListenTLS("tcp", *addr)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
	defer tlsLn.Close()

	// Create an fs.FS from the embedded filesystem
	fSys, err := fs.Sub(site, "webui/public")
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

	// Serve an HTTP file server over TLS using our embedded filesystem
	log.Info("starting file server for web ui")
	err = http.Serve(tlsLn, http.FileServer(http.FS(fSys)))
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}
