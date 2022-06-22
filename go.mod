module project/azure-cosi-driver

go 1.16

require (
	github.com/Azure/azure-storage-blob-go v0.13.0
	github.com/Azure/go-autorest/autorest/to v0.4.0
	google.golang.org/grpc v1.38.0
	k8s.io/client-go v0.22.0
	k8s.io/klog v1.0.0
	k8s.io/klog/v2 v2.10.0
	sigs.k8s.io/cloud-provider-azure v1.1.1
	sigs.k8s.io/container-object-storage-interface-spec v0.0.0-20210507203703-a97f2e98ac90
)
