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
	staticFS embed.FS
	//go:embed templates
	templateFS embed.FS

	addr        = flag.String("addr", ":8080", "the address to listen on in the case of a local listener")
	hostname    = flag.String("hostname", "libations", "hostname to use on the tailnet")
	local       = flag.Bool("local", false, "start on local addr; don't attach to a tailnet")
	recipesFile = flag.String("recipes-file", "", "path to a file containing drink recipes")
	tsnetLogs   = flag.Bool("tsnet-logs", true, "include tsnet logs in application logs")
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

// parseTemplates is used to parse templates from the embedded FS, and ensure that the
// 'StringJoin' function is available to the templates.
func parseTemplates() *template.Template {
	// Create an fs.FS from the embedded filesystem
	files, err := fs.Sub(templateFS, "templates")
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	funcMap := template.FuncMap{"StringsJoin": strings.Join}
	tmpl := template.New("").Funcs(funcMap)
	tmpl, _ = tmpl.ParseFS(files, "*.html", "icons/*.svg")
	return tmpl
}

// parseRecipes attempts to read and parse recipes either from a path specified at the CLI,
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
		if recipesFileContent, err = staticFS.ReadFile("static/sample.json"); err != nil {
			return nil, err
		}

		slog.Info("using recipes from embedded filesystem")
	}

	if err = json.Unmarshal(recipesFileContent, &recipes); err != nil {
		return nil, err
	}

	slog.Info(fmt.Sprintf("loaded %d recipes", len(recipes)))

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

		err := templates.Lookup("index.html").Execute(w, data)
		if err != nil {
			slog.Error("failed to execute template 'index.html'", "error", err.Error())
		}
	})

	return mux
}

// localListener sets up a local TCP listener on the specified addr.
func localListener(addr string) (*net.Listener, error) {
	a, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, err
	}

	httpLn, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1%s", a))
	if err != nil {
		return nil, err
	}

	return &httpLn, nil
}

// tailscaleListener sets up HTTP(s) listeners on the tailnet.
func tailscaleListener() (*net.Listener, error) {
	tsLogger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{}))

	tsnetServer := &tsnet.Server{
		Hostname: *hostname,
		Logf: func(msg string, args ...any) {
			l := tsLogger.With(slog.String("source", "tsnet"), slog.String("hostname", *hostname))
			l.Info(fmt.Sprintf(msg, args...))
		},
	}

	if !*tsnetLogs {
		tsnetServer.Logf = func(string, ...any) {}
		slog.Warn("tsnet logs are disabled, interactive auth link will not be shown")
	}

	// Start a standard HTTP server in the background to redirect HTTP -> HTTPS.
	go func() {
		httpLn, err := tsnetServer.Listen("tcp", ":80")
		if err != nil {
			slog.Error("unable to start HTTP listener, redirects from http->https will not work")
			return
		}

		slog.Info(fmt.Sprintf("started HTTP listener with tsnet at %s:80", *hostname))

		err = http.Serve(httpLn, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			newURL := fmt.Sprintf("https://%s%s", r.Host, r.RequestURI)
			http.Redirect(w, r, newURL, http.StatusMovedPermanently)
		}))
		if err != nil {
			slog.Error("unable to start http server, redirects from http->https will not work")
		}
	}()

	tlsLn, err := tsnetServer.ListenTLS("tcp", ":443")
	if err != nil {
		return nil, err
	}

	return &tlsLn, nil
}

func main() {
	flag.Parse()
	log := slog.Default().With(slog.String("source", "libations"))
	slog.SetDefault(log)

	// Create an fs.FS from the embedded filesystem
	files, err := fs.Sub(staticFS, "static")
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	drinks, err := parseRecipes()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	var listener *net.Listener
	if *local {
		listener, err = localListener(*addr)
	} else {
		listener, err = tailscaleListener()
	}

	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	tmpl := parseTemplates()
	mux := libationsMux(drinks, files, tmpl)

	slog.Info(fmt.Sprintf("starting listener on %s", *addr))
	if err = http.Serve(*listener, mux); err != nil {
		slog.Error(err.Error())
	}
}
