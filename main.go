// Package main implements vice-default-backend, the catch-all backend for
// VICE analysis subdomains. It serves a loading page that periodically
// refreshes until the analysis-specific HTTPRoute takes over and the
// vice-operator loading page (or the analysis itself) responds.
package main

import (
	"embed"
	"flag"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/cyverse-de/app-exposer/common"
	"github.com/sirupsen/logrus"
)

var log = common.Log

//go:embed templates/waiting.html
var waitingTemplateFS embed.FS

// waitingTemplate is the parsed waiting page template.
var waitingTemplate = template.Must(template.ParseFS(waitingTemplateFS, "templates/waiting.html"))

// waitingPageData holds the template data for the waiting page.
type waitingPageData struct {
	// RefreshSeconds is the interval between page reloads.
	RefreshSeconds int
}

// App contains the HTTP handlers for the default backend.
type App struct {
	refreshSeconds int
}

// HandleWaiting serves the waiting page. The page periodically reloads itself;
// once the analysis-specific HTTPRoute is active, the reload lands on the
// vice-operator loading page or the running analysis instead of here.
func (a *App) HandleWaiting(w http.ResponseWriter, r *http.Request) {
	data := waitingPageData{
		RefreshSeconds: a.refreshSeconds,
	}

	var buf strings.Builder
	if err := waitingTemplate.Execute(&buf, data); err != nil {
		log.Errorf("rendering waiting page: %v", err)
		http.Error(w, "failed to render waiting page", http.StatusInternalServerError)
		return
	}

	// Set a custom header so the client-side JS can detect when the response
	// is no longer coming from the default backend.
	w.Header().Set("X-Vice-Default-Backend", "true")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = fmt.Fprint(w, buf.String())
}

func main() {
	log.Logger.SetReportCaller(true)

	var (
		listenAddr     = flag.String("listen", "0.0.0.0:60000", "The listen address.")
		refreshSeconds = flag.Int("refresh-seconds", 5, "Seconds between page reloads while waiting for the analysis route.")
		logLevel       = flag.String("log-level", "info", "One of trace, debug, info, warn, error, fatal, or panic.")
	)

	flag.Parse()

	level, err := logrus.ParseLevel(*logLevel)
	if err != nil {
		log.Fatalf("invalid log level %q: %v", *logLevel, err)
	}
	logrus.SetLevel(level)
	logrus.SetFormatter(&logrus.JSONFormatter{})

	app := &App{
		refreshSeconds: *refreshSeconds,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprint(w, "healthy")
	})
	mux.HandleFunc("/", app.HandleWaiting)

	log.Infof("vice-default-backend listening on %s (refresh-seconds=%d)", *listenAddr, *refreshSeconds)

	server := &http.Server{
		Handler: mux,
		Addr:    *listenAddr,
	}
	log.Fatal(server.ListenAndServe())
}
