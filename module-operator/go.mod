module github.com/verrazzano/verrazzano-modules/module-operator

go 1.19

require (
	github.com/verrazzano/verrazzano 43908c21553a351c224669f5ae2affc9d66328b9
	github.com/Jeffail/gabs/v2 v2.6.1
	github.com/cert-manager/cert-manager v1.9.1
	github.com/crossplane/crossplane-runtime v0.17.0
	github.com/crossplane/oam-kubernetes-runtime v0.3.3
	github.com/gertd/go-pluralize v0.2.0
	github.com/go-logr/logr v1.2.3
	github.com/golang/mock v1.6.0
	github.com/google/go-cmp v0.5.9
	github.com/google/uuid v1.3.0
	github.com/gordonklaus/ineffassign v0.0.0-20210104184537-8eed68eb605f
	github.com/hashicorp/go-retryablehttp v0.6.8
	github.com/mattn/go-isatty v0.0.14
	github.com/onsi/ginkgo/v2 v2.1.6
	github.com/onsi/gomega v1.20.1
	github.com/oracle/oci-go-sdk/v53 v53.1.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring v0.59.1
	github.com/prometheus/client_golang v1.12.1
	github.com/spf13/cobra v1.6.1
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.8.0
	github.com/verrazzano/verrazzano-monitoring-operator v0.0.31-0.20230201202534-ac5ebe880e95
	go.uber.org/zap v1.21.0
	golang.org/x/lint v0.0.0-20210508222113-6edffad5e616
	golang.org/x/text v0.4.0
	golang.org/x/tools v0.1.12
	google.golang.org/protobuf v1.28.1
	gopkg.in/yaml.v3 v3.0.1
	helm.sh/helm/v3 v3.10.3
	istio.io/api v0.0.0-20221208152505-d807bc07da6a
	istio.io/client-go v1.15.4
	k8s.io/api v0.25.2
	k8s.io/apiextensions-apiserver v0.25.2
	k8s.io/apimachinery v0.25.2
	k8s.io/cli-runtime v0.25.2
	k8s.io/client-go v0.25.2
	k8s.io/code-generator v0.25.2
	sigs.k8s.io/cluster-api v1.2.0
	sigs.k8s.io/controller-runtime v0.12.3
	sigs.k8s.io/controller-tools v0.9.2
	sigs.k8s.io/yaml v1.3.0
)



replace (
	github.com/ajeddeloh/go-json v0.0.0-20200220154158-5ae607161559 => github.com/coreos/go-json v0.0.0-20220325222439-31b2177291ae
	github.com/crossplane/crossplane-runtime => github.com/verrazzano/crossplane-runtime v0.17.0-1
	github.com/crossplane/oam-kubernetes-runtime => github.com/verrazzano/oam-kubernetes-runtime v0.3.3-3
	github.com/emicklei/go-restful => github.com/emicklei/go-restful v2.16.0+incompatible
	github.com/gogo/protobuf => github.com/gogo/protobuf v1.3.2
	github.com/onsi/ginkgo/v2 => github.com/onsi/ginkgo/v2 v2.0.0
	github.com/onsi/gomega => github.com/onsi/gomega v1.17.0
	github.com/spf13/cobra => github.com/spf13/cobra v1.6.1
	github.com/stretchr/testify => github.com/stretchr/testify v1.7.1
	golang.org/x/lint => golang.org/x/lint v0.0.0-20201208152925-83fdc39ff7b5
	gopkg.in/yaml.v3 => gopkg.in/yaml.v3 v3.0.1
	helm.sh/helm/v3 => helm.sh/helm/v3 v3.10.3
	sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.11.2
	sigs.k8s.io/controller-tools => sigs.k8s.io/controller-tools v0.8.0
	sigs.k8s.io/kind => github.com/verrazzano/kind v0.0.0-20221129215948-885481909133
)
