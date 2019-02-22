package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"bou.ke/kubectl-dashboard/files"

	"github.com/kubernetes/dashboard/src/app/backend/args"
	"github.com/kubernetes/dashboard/src/app/backend/handler"
	"github.com/kubernetes/dashboard/src/app/backend/integration"
	"github.com/kubernetes/dashboard/src/app/backend/settings"
	"github.com/kubernetes/dashboard/src/app/backend/systembanner"
	"github.com/pkg/browser"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/pflag"
)

func main() {
	log.SetOutput(os.Stdout)

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	flag.CommandLine.Parse(make([]string, 0))

	l, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		log.Fatal(err)
	}

	initArgHolder()

	cm := newClientManager()

	settingsManager := settings.NewSettingsManager(cm)
	systemBannerManager := systembanner.NewSystemBannerManager("", "")
	integrationManager := integration.NewIntegrationManager(cm)

	apiHandler, err := handler.CreateHTTPAPIHandler(
		integrationManager,
		cm,
		&authManager{},
		settingsManager,
		systemBannerManager)
	if err != nil {
		log.Fatal(err)
	}

	http.Handle("/", files.Server)
	http.Handle("/api/", apiHandler)
	http.Handle("/config", handler.AppHandler(handler.ConfigHandler))
	http.Handle("/api/sockjs/", handler.CreateAttachHandler("/api/sockjs"))
	http.Handle("/metrics", prometheus.Handler())

	log.Printf("Serving at http://%v/", l.Addr())
	if browser.OpenURL(fmt.Sprintf("http://%s/", l.Addr())) == nil {
		log.Print("Opening browser...")
	}
	log.Fatal(http.Serve(l, http.DefaultServeMux))
}

func initArgHolder() {
	builder := args.GetHolderBuilder()
	builder.SetEnableInsecureLogin(true)
	builder.SetAPILogLevel("INFO")
}
