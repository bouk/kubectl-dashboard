module bou.ke/kubectl-dashboard

go 1.14

require (
	github.com/emicklei/go-restful v2.12.0+incompatible
	github.com/kubernetes/dashboard v1.10.1
	github.com/pkg/browser v0.0.0-20180916011732-0a3d74bf9ce4
	github.com/spf13/pflag v1.0.5
	k8s.io/api v0.18.3
	k8s.io/apiextensions-apiserver v0.18.2
	k8s.io/client-go v0.18.3
)

replace github.com/kubernetes/dashboard => github.com/bouk/dashboard v1.10.1-0.20200521183814-5b83803463d0
