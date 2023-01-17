module github.com/kubeedge/mappers-go

go 1.14

require (
	github.com/beevik/etree v1.1.0
	github.com/currantlabs/ble v0.0.0-20171229162446-c1d21c164cf8
	github.com/eclipse/paho.mqtt.golang v1.3.0
	github.com/goburrow/modbus v0.1.0
	github.com/goburrow/serial v0.1.0
	github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e // indirect
	github.com/gopcua/opcua v0.2.6
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/kubeedge/kubeedge v1.5.0
	github.com/mgutz/ansi v0.0.0-20200706080929-d51e80ef957d // indirect
	github.com/mgutz/logxi v0.0.0-20161027140823-aebf8a7d67ab // indirect
	github.com/onsi/ginkgo v1.15.0
	github.com/onsi/gomega v1.10.5
	github.com/pkg/errors v0.9.1
	github.com/sailorvii/goav v0.1.4
	github.com/sailorvii/modbus v0.1.2
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.6.1
	github.com/tbrandon/mbserver v0.0.0-20210320091329-a1f8ae952881
	github.com/use-go/onvif v0.0.1
	golang.org/x/net v0.0.0-20201202161906-c7110b5ffcbb
	google.golang.org/protobuf v1.25.0 // indirect
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.19.3
	k8s.io/klog v1.0.0
	k8s.io/klog/v2 v2.4.0
)

replace (
	github.com/kubeedge/beehive v0.0.0 => github.com/kubeedge/beehive v0.0.0-20201125122335-cd19bca6e436
	github.com/kubeedge/viaduct v0.0.0 => github.com/kubeedge/viaduct v0.0.0-20201130063818-e33931917980
	k8s.io/api v0.0.0 => k8s.io/api v0.19.3
	k8s.io/apiextensions-apiserver v0.0.0 => k8s.io/apiextensions-apiserver v0.19.3
	k8s.io/apimachinery v0.0.0 => k8s.io/apimachinery v0.19.3
	k8s.io/apiserver v0.0.0 => k8s.io/apiserver v0.19.3
	k8s.io/cli-runtime v0.0.0 => k8s.io/cli-runtime v0.19.3
	k8s.io/client-go v0.0.0 => k8s.io/client-go v0.19.3
	k8s.io/cloud-provider v0.0.0 => k8s.io/cloud-provider v0.19.3
	k8s.io/cluster-bootstrap v0.0.0 => k8s.io/cluster-bootstrap v0.19.3
	k8s.io/code-generator v0.0.0 => k8s.io/code-generator v0.19.3
	k8s.io/component-base v0.0.0 => k8s.io/component-base v0.19.3
	k8s.io/cri-api v0.0.0 => k8s.io/cri-api v0.19.3
	k8s.io/csi-api v0.0.0 => k8s.io/csi-api v0.19.3
	k8s.io/csi-translation-lib v0.0.0 => k8s.io/csi-translation-lib v0.19.3
	k8s.io/gengo v0.0.0 => k8s.io/gengo v0.19.3
	k8s.io/heapster => k8s.io/heapster v1.2.0-beta.1 // indirect
	k8s.io/klog/v2 => k8s.io/klog/v2 v2.2.0
	k8s.io/kube-aggregator v0.0.0 => k8s.io/kube-aggregator v0.19.3
	k8s.io/kube-controller-manager v0.0.0 => k8s.io/kube-controller-manager v0.19.3
	k8s.io/kube-openapi v0.0.0 => k8s.io/kube-openapi v0.19.3
	k8s.io/kube-proxy v0.0.0 => k8s.io/kube-proxy v0.19.3
	k8s.io/kube-scheduler v0.0.0 => k8s.io/kube-scheduler v0.19.3
	k8s.io/kubectl => k8s.io/kubectl v0.19.3
	k8s.io/kubelet v0.0.0 => k8s.io/kubelet v0.19.3
	k8s.io/legacy-cloud-providers v0.0.0 => k8s.io/legacy-cloud-providers v0.19.3
	k8s.io/metrics v0.0.0 => k8s.io/metrics v0.19.3
	k8s.io/node-api v0.0.0 => k8s.io/node-api v0.19.3
	k8s.io/repo-infra v0.0.0 => k8s.io/repo-infra v0.19.3
	k8s.io/sample-apiserver v0.0.0 => k8s.io/sample-apiserver v0.19.3
	k8s.io/utils v0.0.0 => k8s.io/utils v0.19.3
)
