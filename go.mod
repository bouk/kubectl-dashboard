module bou.ke/kubectl-dashboard

go 1.14

require (
	github.com/emicklei/go-restful v2.13.0+incompatible
	github.com/kubernetes/dashboard v1.10.1
	github.com/pkg/browser v0.0.0-20180916011732-0a3d74bf9ce4
	github.com/spf13/pflag v1.0.5
	k8s.io/api v0.18.4
	k8s.io/apiextensions-apiserver v0.18.4
	k8s.io/client-go v0.18.4
)

replace github.com/kubernetes/dashboard => github.com/kubernetes/dashboard v1.10.1-0.20200622141346-665b4d367d61
