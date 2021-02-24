module github.com/form3tech-oss/f1

//test
go 1.14

require (
	github.com/aholic/ggtimer v0.0.0-20150905131044-5d7b30837a52
	github.com/asaskevich/govalidator v0.0.0-20200907205600-7a23bdc65eef
	github.com/blend/go-sdk v1.20210222.1 // indirect
	github.com/bmizerany/assert v0.0.0-20160611221934-b7ed37b82869 // indirect
	github.com/chobie/go-gaussian v0.0.0-20150107165016-53c09d90eeaf
	github.com/evalphobia/logrus_fluent v0.5.4
	github.com/fluent/fluent-logger-golang v1.5.0 // indirect
	github.com/giantswarm/retry-go v0.0.0-20151203102909-d78cea247d5e
	github.com/golang/freetype v0.0.0-20170609003504-e2365dfdc4a0 // indirect
	github.com/google/uuid v1.2.0
	github.com/guptarohit/asciigraph v0.5.1
	github.com/hako/durafmt v0.0.0-20200710122514-c0fb7b4da026
	github.com/juju/errgo v0.0.0-20140925100237-08cceb5d0b53 // indirect
	github.com/magefile/mage v1.11.0 // indirect
	github.com/phayes/freeport v0.0.0-20180830031419-95f893ade6f2
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.9.0
	github.com/prometheus/client_model v0.2.0
	github.com/prometheus/common v0.17.0
	github.com/prometheus/procfs v0.6.0 // indirect
	github.com/sirupsen/logrus v1.8.0
	github.com/spf13/cobra v1.1.3
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.7.0
	github.com/tinylib/msgp v1.1.5 // indirect
	github.com/wcharczuk/go-chart v2.0.2-0.20191206192251-962b9abdec2b+incompatible
	github.com/workanator/go-ataman v0.0.0-20201223053433-503c6ff9de7d
	go.uber.org/goleak v1.1.10
	golang.org/x/image v0.0.0-20210220032944-ac19c3e999fb // indirect
	golang.org/x/lint v0.0.0-20201208152925-83fdc39ff7b5 // indirect
	golang.org/x/sys v0.0.0-20210223212115-eede4237b368 // indirect
	golang.org/x/tools v0.1.0 // indirect
	gopkg.in/workanator/go-ataman.v1 v1.0.0-20201223053604-e3b73d2e8108 // indirect
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
)

replace (
	github.com/Sirupsen/logrus => github.com/sirupsen/logrus v1.6.0
	github.com/ahmetb/go-linq => github.com/ahmetb/go-linq/v3 v3.1.0
	github.com/docker/distribution => github.com/docker/distribution v0.0.0-20200319173657-742aab907b54
	github.com/docker/docker => github.com/docker/engine v0.0.0-20190725163905-fa8dd90ceb7b
	github.com/docker/libcompose => github.com/docker/libcompose v0.4.1-0.20190808084053-143e0f3f1ab9
	github.com/form3tech-oss/go-vault-client => github.com/form3tech-oss/go-vault-client v2.0.1+incompatible
	github.com/opencontainers/runc => github.com/opencontainers/runc v1.0.0-rc3
	github.com/optiopay/kafka => github.com/optiopay/kafka/v2 v2.1.1
	k8s.io/api => k8s.io/api v0.15.11
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.15.11
	k8s.io/apimachinery => k8s.io/apimachinery v0.15.11
	k8s.io/apiserver => k8s.io/apiserver v0.15.11
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.15.11
	k8s.io/client-go => k8s.io/client-go v0.15.11
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.15.11
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.15.11
	k8s.io/code-generator => k8s.io/code-generator v0.15.11
	k8s.io/component-base => k8s.io/component-base v0.15.11
	k8s.io/cri-api => k8s.io/cri-api v0.15.11
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.15.11
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.15.11
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.15.11
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.15.11
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.15.11
	k8s.io/kubectl => k8s.io/kubectl v0.15.11
	k8s.io/kubelet => k8s.io/kubelet v0.15.11
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.15.11
	k8s.io/metrics => k8s.io/metrics v0.15.11
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.15.11
)
