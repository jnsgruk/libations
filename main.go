package main

//go:generate bash -c "hugo -s ./webui --minify"

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"os"

	"tailscale.com/tsnet"
)

var (
	//go:embed webui/public
	site embed.FS

	hostname  = flag.String("hostname", "libations", "hostname to use on the tailnet")
	tsnetLogs = flag.Bool("tsnet-logs", true, "include tsnet logs in application logs")
	local     = flag.Bool("local", false, "start on local addr; don't attach to a tailnet")
	addr      = flag.String("addr", ":8080", "the address to listen on in the case of a local listener")
)

// redirectToTLS is a simple http hander that redirects all HTTP requests to HTTPs.
func redirectToTLS(w http.ResponseWriter, r *http.Request) {
	newURL := fmt.Sprintf("https://%s%s", r.Host, r.RequestURI)
	http.Redirect(w, r, newURL, http.StatusMovedPermanently)
}

// serveLocal sets up a local TCP listener on the specified addr, and then serves the embedded
// filesystem over HTTP on the given listener.
func serveLocal(files fs.FS, addr string) {
	a, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	httpLn, err := net.ListenTCP("tcp", a)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	slog.Info(fmt.Sprintf("started HTTP listener on %s", addr))

	// Serve an HTTP file server using our embedded filesystem
	slog.Info("starting file server for web ui")
	err = http.Serve(httpLn, http.FileServer(http.FS(files)))
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}

// serveTailscale sets up HTTP & HTTPs listeners on the tailnet and serves the embedded FS on them.
func serveTailscale(files fs.FS) {
	tsLogger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{}))

	tsnetServer := &tsnet.Server{
		Hostname: *hostname,
		Logf: func(msg string, args ...any) {
			l := tsLogger.With(slog.String("source", "tsnet"), slog.String("hostname", *hostname))
			l.Info(fmt.Sprintf(msg, args...))
		},
	}
	defer tsnetServer.Close()

	if !*tsnetLogs {
		tsnetServer.Logf = func(string, ...any) {}
		slog.Warn("tsnet logs are disabled, interactive auth link will not be shown")
	}

	// Start a standard HTTP server in the background to redirect HTTP -> HTTPS
	go func() {
		httpLn, err := tsnetServer.Listen("tcp", ":80")
		if err != nil {
			slog.Error(err.Error())
			os.Exit(1)
		}

		slog.Info(fmt.Sprintf("started HTTP listener with tsnet at %s:80", *hostname))

		err = http.Serve(httpLn, http.HandlerFunc(redirectToTLS))
		if err != nil {
			slog.Error(err.Error())
			os.Exit(1)
		}
	}()

	tlsLn, err := tsnetServer.ListenTLS("tcp", ":443")
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	defer tlsLn.Close()

	slog.Info(fmt.Sprintf("started HTTPS listener with tsnet at %s:443", *hostname))

	// Serve an HTTP file server over TLS using our embedded filesystem
	slog.Info("starting file server for web ui")
	err = http.Serve(tlsLn, http.FileServer(http.FS(files)))
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}

func main() {
	flag.Parse()

	// Configure the default logger
	log := slog.Default().With(slog.String("source", "libations"))
	slog.SetDefault(log)

	// Create an fs.FS from the embedded filesystem
	files, err := fs.Sub(site, "webui/public")
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

	if *local {
		serveLocal(files, *addr)
	} else {
		serveTailscale(files)
	}
}
