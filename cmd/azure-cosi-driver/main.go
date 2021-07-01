// Copyright 2021 The Ceph-COSI Authors.
// Licensed under the Apache License, Version 2.0 (the "License");
// You may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"flag"
	"project/azure-cosi-driver/pkg/driver"
	identityserver "project/azure-cosi-driver/pkg/server/identity"
	provisionerserver "project/azure-cosi-driver/pkg/server/provisioner"

	"k8s.io/klog"
)

func main() {
	klog.InitFlags(nil)
	endpoint := flag.String("endpoint", driver.DefaultEndpoint, "endpoint for the GRPC server")
	flag.Parse()

	defer klog.Flush()

	provServer := provisionerserver.NewProvisionerServer()
	identityServer := identityserver.NewIdentityServer(driver.DriverName)

	err := driver.Run(*endpoint, identityServer, provServer)
	if err != nil {
		klog.Exitf("Error when running driver: %v", err)
	}
}
