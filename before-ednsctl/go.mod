module github.com/lucasreed/go-interface-refactoring/before-ednsctl

go 1.12

require (
	github.com/aws/aws-sdk-go v1.25.9
	github.com/cloudflare/cloudflare-go v0.10.3
	github.com/davecgh/go-spew v1.1.1
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79 // indirect
	github.com/lithammer/dedent v1.1.0
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
	github.com/spf13/cobra v0.0.5
	google.golang.org/api v0.11.0
	k8s.io/api v0.0.0-20191005115622-2e41325d9e4b
	k8s.io/apimachinery v0.0.0-20191005115455-e71eb83a557c
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/klog v1.0.0
	sigs.k8s.io/controller-runtime v0.2.2
)

replace (
	k8s.io/api => k8s.io/api v0.0.0-20190111032252-67edc246be36
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20181127025237-2b1284ed4c93
	k8s.io/client-go => k8s.io/client-go v0.0.0-20190111032708-6bf63545bd02
)
