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
		nil,
		settingsManager,
		systemBannerManager)
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", files.Server)
	mux.Handle("/api/", apiHandler)
	mux.Handle("/config", handler.AppHandler(handler.ConfigHandler))
	mux.Handle("/api/sockjs/", handler.CreateAttachHandler("/api/sockjs"))

	log.Printf("Serving at http://%v/", l.Addr())
	if browser.OpenURL(fmt.Sprintf("http://%s/", l.Addr())) == nil {
		log.Print("Opening browser...")
	}
	log.Fatal(http.Serve(l, mux))
}

func initArgHolder() {
	builder := args.GetHolderBuilder()
	builder.SetAPILogLevel("INFO")
}
