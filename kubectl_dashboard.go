package main

import (
	"flag"
	"log"
	"net"
	"net/http"

	"github.com/kubernetes/dashboard/src/app/backend/auth"
	authApi "github.com/kubernetes/dashboard/src/app/backend/auth/api"
	"github.com/kubernetes/dashboard/src/app/backend/auth/jwe"
	"github.com/kubernetes/dashboard/src/app/backend/client"
	clientapi "github.com/kubernetes/dashboard/src/app/backend/client/api"
	"github.com/kubernetes/dashboard/src/app/backend/handler"
	"github.com/kubernetes/dashboard/src/app/backend/integration"
	"github.com/kubernetes/dashboard/src/app/backend/settings"
	"github.com/kubernetes/dashboard/src/app/backend/sync"
	"github.com/kubernetes/dashboard/src/app/backend/systembanner"
	"github.com/prometheus/client_golang/prometheus"
)

func main() {
	flag.Parse()

	clientManager := client.NewClientManager("/home/bouke/.kube/config", "")
	versionInfo, err := clientManager.InsecureClient().Discovery().ServerVersion()
	if err != nil {
		handleFatalInitError(err)
	}

	log.Printf("Successful initial request to the apiserver, version: %s", versionInfo.String())

	// Init auth manager
	authManager := initAuthManager(clientManager)

	// Init settings manager
	settingsManager := settings.NewSettingsManager(clientManager)

	// Init system banner manager
	systemBannerManager := systembanner.NewSystemBannerManager("", "")

	// Init integrations
	integrationManager := integration.NewIntegrationManager(clientManager)
	//integrationManager.Metric().ConfigureHeapster(args.Holder.GetHeapsterHost()).
	//EnableWithRetry(integrationapi.HeapsterIntegrationID, time.Duration(args.Holder.GetMetricClientCheckPeriod()))

	apiHandler, err := handler.CreateHTTPAPIHandler(
		integrationManager,
		clientManager,
		authManager,
		settingsManager,
		systemBannerManager)
	if err != nil {
		handleFatalInitError(err)
	}

	// Run a HTTP server that serves static public files from './public' and handles API calls.
	// TODO(bryk): Disable directory listing.
	http.Handle("/", handler.MakeGzipHandler(handler.CreateLocaleHandler()))
	http.Handle("/api/", apiHandler)
	// TODO(maciaszczykm): Move to /appConfig.json as it was discussed in #640.
	http.Handle("/api/appConfig.json", handler.AppHandler(handler.ConfigHandler))
	http.Handle("/api/sockjs/", handler.CreateAttachHandler("/api/sockjs"))
	http.Handle("/metrics", prometheus.Handler())

	l, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		handleFatalInitError(err)
	}

	log.Printf("Serving insecurely on HTTP: %v", l.Addr())
	log.Fatal(http.Serve(l, http.DefaultServeMux))
}

func initAuthManager(clientManager clientapi.ClientManager) authApi.AuthManager {
	insecureClient := clientManager.InsecureClient()

	// Init default encryption key synchronizer
	synchronizerManager := sync.NewSynchronizerManager(insecureClient)
	keySynchronizer := synchronizerManager.Secret(authApi.EncryptionKeyHolderNamespace, authApi.EncryptionKeyHolderName)

	// Register synchronizer. Overwatch will be responsible for restarting it in case of error.
	sync.Overwatch.RegisterSynchronizer(keySynchronizer, sync.AlwaysRestart)

	// Init encryption key holder and token manager
	keyHolder := jwe.NewRSAKeyHolder(keySynchronizer)
	tokenManager := jwe.NewJWETokenManager(keyHolder)

	// Set token manager for client manager.
	clientManager.SetTokenManager(tokenManager)
	authModes := authApi.AuthenticationModes{}
	authModes.Add(authApi.Token)

	// UI logic dictates this should be the inverse of the cli option
	authenticationSkippable := true

	return auth.NewAuthManager(clientManager, tokenManager, authModes, authenticationSkippable)
}

/**
 * Handles fatal init error that prevents server from doing any work. Prints verbose error
 * message and quits the server.
 */
func handleFatalInitError(err error) {
	log.Fatalf("Error while initializing connection to Kubernetes apiserver. "+
		"This most likely means that the cluster is misconfigured (e.g., it has "+
		"invalid apiserver certificates or service account's configuration) or the "+
		"--apiserver-host param points to a server that does not exist. Reason: %s\n"+
		"Refer to our FAQ and wiki pages for more information: "+
		"https://github.com/kubernetes/dashboard/wiki/FAQ", err)
}
