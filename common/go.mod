module github.com/verrazzano/verrazzano-modules/common

go 1.19

require (
	github.com/verrazzano/verrazzano v1.5.2
	go.uber.org/zap v1.24.0
	k8s.io/apimachinery v0.27.0
	sigs.k8s.io/controller-runtime v0.14.6
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/emicklei/go-restful/v3 v3.9.0 // indirect
	github.com/evanphx/json-patch v5.6.0+incompatible // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/go-logr/zapr v1.2.3 // indirect
	github.com/go-openapi/jsonpointer v0.19.6 // indirect
	github.com/go-openapi/jsonreference v0.20.1 // indirect
	github.com/go-openapi/swag v0.22.3 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/gnostic v0.6.9 // indirect
	github.com/google/go-cmp v0.5.9 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/imdario/mergo v0.3.12 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/client_golang v1.14.0 // indirect
	github.com/prometheus/client_model v0.3.0 // indirect
	github.com/prometheus/common v0.37.0 // indirect
	github.com/prometheus/procfs v0.8.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/stretchr/testify v1.8.2 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	golang.org/x/net v0.8.0 // indirect
	golang.org/x/oauth2 v0.0.0-20220909003341-f21342109be1 // indirect
	golang.org/x/sys v0.6.0 // indirect
	golang.org/x/term v0.6.0 // indirect
	golang.org/x/text v0.8.0 // indirect
	golang.org/x/time v0.3.0 // indirect
	gomodules.xyz/jsonpatch/v2 v2.2.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/protobuf v1.28.1 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/api v0.27.0 // indirect
	k8s.io/apiextensions-apiserver v0.27.0 // indirect
	k8s.io/client-go v0.27.0 // indirect
	k8s.io/component-base v0.27.0 // indirect
	k8s.io/klog/v2 v2.90.1 // indirect
	k8s.io/kube-openapi v0.0.0-20230308215209-15aac26d736a // indirect
	k8s.io/utils v0.0.0-20230209194617-a36077c30491 // indirect
	sigs.k8s.io/json v0.0.0-20221116044647-bc3834ca7abd // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.3 // indirect
	sigs.k8s.io/yaml v1.3.0 // indirect
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
	k8s.io/api => k8s.io/api v0.25.2
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.25.2
	k8s.io/apimachinery => k8s.io/apimachinery v0.25.2
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.25.2
	k8s.io/client-go => k8s.io/client-go v0.25.2
	k8s.io/code-generator => k8s.io/code-generator v0.25.2
	sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.11.2
	sigs.k8s.io/controller-tools => sigs.k8s.io/controller-tools v0.8.0
	sigs.k8s.io/kind => github.com/verrazzano/kind v0.0.0-20221129215948-885481909133
)
