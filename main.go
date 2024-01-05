package main

import (
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"tailscale.com/tsnet"
)

var (
	//go:embed static
	site embed.FS

	hostname    = flag.String("hostname", "libations", "hostname to use on the tailnet")
	tsnetLogs   = flag.Bool("tsnet-logs", true, "include tsnet logs in application logs")
	local       = flag.Bool("local", false, "start on local addr; don't attach to a tailnet")
	addr        = flag.String("addr", ":8080", "the address to listen on in the case of a local listener")
	recipesFile = flag.String("recipes-file", "", "path to a file containing drink recipes")
)

// Ingredient represents the name and quantity of a given ingredient in a recipe.
type Ingredient struct {
	Name    string
	Measure string
	Unit    string
}

// Drink represents all of the details for a given drink.
type Drink struct {
	ID          int
	Name        string
	Base        []string
	Glass       []string
	Method      []string
	Ice         []string
	Ingredients []Ingredient
	Garnish     []string
	Notes       string
}

// LibationsPageData is used for rendering the web page templates with the relevant information.
type LibationsPageData struct {
	Time   string
	Drinks []Drink
}

// parseTemplates is used to parse templates from various directories, and ensure that the
// 'StringJoin' function is available to the templates.
func parseTemplates() *template.Template {
	funcMap := template.FuncMap{"StringsJoin": strings.Join}
	tmpl := template.New("").Funcs(funcMap)
	globs := []string{"templates/*.html", "templates/icons/*.svg"}

	// Iterate over each of the globbed paths, adding templates.
	for _, g := range globs {
		if t, _ := tmpl.ParseGlob(g); t != nil {
			tmpl = t
		}
	}

	return tmpl
}

// parseRecipes attempts to read and parse recipes either from a path specifed at the CLI,
// or the default set of recipes included in the embedded filesystem.
func parseRecipes() ([]Drink, error) {
	var err error
	var recipes []Drink
	var recipesFileContent []byte

	if *recipesFile != "" {
		if recipesFileContent, err = os.ReadFile(*recipesFile); err != nil {
			return nil, err
		}

		slog.Info(fmt.Sprintf("using recipes file at: %s", *recipesFile))
	} else {
		if recipesFileContent, err = site.ReadFile("static/sample.json"); err != nil {
			return nil, err
		}

		slog.Info("using recipes from embedded filesystem")
	}

	if err = json.Unmarshal(recipesFileContent, &recipes); err != nil {
		return nil, err
	}

	return recipes, nil
}

// libationsMux returns an http.ServeMux that knows to to handle the routes required for the app.
func libationsMux(drinks []Drink, files fs.FS, templates *template.Template) *http.ServeMux {
	mux := http.NewServeMux()

	// Serve static files from our embedded filesystem using http.Fileserver
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(files))))

	// Render the templates with the drinks/time data.
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data := LibationsPageData{
			Time:   time.Now().Format("January 2, 2006 at 15:04 MST"),
			Drinks: drinks,
		}
		templates.Lookup("index.html").Execute(w, data)
	})

	return mux
}

// redirectToTLS is a simple http hander that redirects all HTTP requests to HTTPs.
func redirectToTLS(w http.ResponseWriter, r *http.Request) {
	newURL := fmt.Sprintf("https://%s%s", r.Host, r.RequestURI)
	http.Redirect(w, r, newURL, http.StatusMovedPermanently)
}

// serveLocal sets up a local TCP listener on the specified addr, and then serves the embedded
// site over HTTP on the given listener.
func serveLocal(drinks []Drink, files fs.FS, addr string) {
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

	tmpl := parseTemplates()
	mux := libationsMux(drinks, files, tmpl)

	if err = http.Serve(httpLn, mux); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}

// serveTailscale sets up HTTP(s) listeners on the tailnet and serves the embedded site on them.
func serveTailscale(drinks []Drink, files fs.FS) {
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

	tmpl := parseTemplates()
	mux := libationsMux(drinks, files, tmpl)

	if err = http.Serve(tlsLn, mux); err != nil {
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
	files, err := fs.Sub(site, "static")
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

	drinks, err := parseRecipes()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	if *local {
		serveLocal(drinks, files, *addr)
	} else {
		serveTailscale(drinks, files)
	}
}
