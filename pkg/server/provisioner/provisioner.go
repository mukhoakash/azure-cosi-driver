// Copyright 2021 The Kubernetes Authors.
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

package provisionerserver

import (
	"context"
	"fmt"
	"project/azure-cosi-driver/pkg/azureutils"
	"reflect"
	"sync"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog"
	spec "sigs.k8s.io/container-object-storage-interface-spec"
)

type bucket struct {
	bucketId   string
	bucketName string
	parameters map[string]string
}

type provisioner struct {
	spec.UnimplementedProvisionerServer

	bucketsLock       sync.RWMutex
	nameToBucketMap   map[string]*bucket
	bucketIdToNameMap map[string]string
}

var _ spec.ProvisionerServer = &provisioner{}

func NewProvisionerServer() spec.ProvisionerServer {
	return &provisioner{
		nameToBucketMap:   make(map[string]*bucket),
		bucketsLock:       sync.RWMutex{},
		bucketIdToNameMap: make(map[string]string),
	}
}

func (pr *provisioner) ProvisionerCreateBucket(
	ctx context.Context,
	req *spec.ProvisionerCreateBucketRequest) (*spec.ProvisionerCreateBucketResponse, error) {
	protocol := req.GetProtocol()
	if protocol == nil {
		return nil, status.Error(codes.InvalidArgument, "Protocol is nil")
	}

	azureBlob := protocol.GetAzureBlob()
	if azureBlob == nil {
		return nil, status.Error(codes.InvalidArgument, "Azure blob protocol is missing")
	}

	storageAccountName := azureBlob.StorageAccount
	bucketName := req.GetName()
	parameters := req.GetParameters()

	if parameters == nil {
		parameters = make(map[string]string)
	}

	// Check if a bucket with these set of values exist in the namesToBucketMap
	pr.bucketsLock.RLock()
	currBucket, exists := pr.nameToBucketMap[bucketName]
	pr.bucketsLock.RUnlock()

	if exists {
		bucketParams := currBucket.parameters
		if bucketParams == nil {
			bucketParams = make(map[string]string)
		}
		if reflect.DeepEqual(bucketParams, parameters) {
			return &spec.ProvisionerCreateBucketResponse{
				BucketId: currBucket.bucketId,
			}, nil
		}

		return nil, status.Error(codes.AlreadyExists, fmt.Sprintf("Bucket %s exists with different parameters", bucketName))
	}

	bucketId, err := azureutils.CreateContainer(ctx, storageAccountName, azureutils.AccessKey, bucketName, parameters)
	if err != nil {
		return nil, err
	}

	// Insert the bucket into the namesToBucketMap
	pr.bucketsLock.RLock()
	pr.nameToBucketMap[bucketName] = &bucket{
		bucketId:   bucketId,
		parameters: parameters,
	}
	pr.bucketIdToNameMap[bucketId] = bucketName
	pr.bucketsLock.RUnlock()

	klog.Infof("ProvisionerCreateBucket :: Bucket id :: %s", bucketId)

	return &spec.ProvisionerCreateBucketResponse{
		BucketId: bucketId,
	}, nil
}

func (pr *provisioner) ProvisionerDeleteBucket(
	ctx context.Context,
	req *spec.ProvisionerDeleteBucketRequest) (*spec.ProvisionerDeleteBucketResponse, error) {
	bucketId := req.GetBucketId()
	klog.Infof("ProvisionerDeleteBucket :: Bucket id :: %s", bucketId)
	err := azureutils.DeleteContainer(ctx, bucketId, azureutils.AccessKey)

	if err != nil {
		return nil, err
	}

	if bucketName, ok := pr.bucketIdToNameMap[bucketId]; ok {
		// Remove from the namesToBucketMap
		pr.bucketsLock.RLock()
		delete(pr.nameToBucketMap, bucketName)
		delete(pr.bucketIdToNameMap, bucketId)
		pr.bucketsLock.RUnlock()
	}

	return &spec.ProvisionerDeleteBucketResponse{}, nil
}

func (pr *provisioner) ProvisionerGrantBucketAccess(
	ctx context.Context,
	req *spec.ProvisionerGrantBucketAccessRequest) (*spec.ProvisionerGrantBucketAccessResponse, error) {
	return &spec.ProvisionerGrantBucketAccessResponse{}, nil
}

func (pr *provisioner) ProvisionerRevokeBucketAccess(
	ctx context.Context,
	req *spec.ProvisionerRevokeBucketAccessRequest) (*spec.ProvisionerRevokeBucketAccessResponse, error) {
	return &spec.ProvisionerRevokeBucketAccessResponse{}, nil
}
