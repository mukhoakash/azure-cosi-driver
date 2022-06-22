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

package azureutils

import (
	"context"
	"fmt"
	"net/url"
	"regexp"

	"github.com/Azure/azure-storage-blob-go/azblob"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	azure "sigs.k8s.io/cloud-provider-azure/pkg/provider"
)

const (
	AccessKey = "t022Cjyk6NsuJsq0s+rBoTs3YQM/FjM2akIs/WIUah9dwz7AD+NiKr/QqQe9nzs8/aXeceH/NtbsK6wyAunr6A=="
)

var (
	storageAccountRE = regexp.MustCompile(`https://(.+).blob.core.windows.net/([^/]*)/?(.*)`)
)

func CreateBucket(
	ctx context.Context,
	storageAccount string,
	containerName string,
	parameters map[string]string,
	cloud *azure.Cloud) (string, error) {

	options, err := parseParametersForStorageAccount(parameters, cloud)
	if err != nil {
		return "", status.Error(codes.Unknown, fmt.Sprintf("Error parsing parameters : %v", err))
	}

	options.Name = storageAccount
	options.ResourceGroup = cloud.ResourceGroup
	options.CreateAccount = true

	// Check and create StorageAccount here
	accountName, accessKey, err := cloud.EnsureStorageAccount(options, "")
	if err != nil {
		return "", status.Error(codes.Unknown, fmt.Sprintf("Error creating storage account with name %s : %v", storageAccount, err))
	}

	// Once storage account is created, we create the azure container inside the storage account
	return createAzureContainer(ctx, accountName, accessKey, containerName, parameters)
}

func DeleteBucket(
	ctx context.Context,
	bucketId string,
	cloud *azure.Cloud) error {
	// Get storage account name from bucketId
	storageAccountName := getStorageAccountNameFromContainerUrl(bucketId)
	// Get access keys for the storage account
	accessKey, err := cloud.GetStorageAccesskey(storageAccountName, cloud.ResourceGroup)
	if err != nil {
		return err
	}

	containerName := getContainerNameFromContainerUrl(bucketId)
	err = deleteAzureContainer(ctx, storageAccountName, accessKey, containerName)
	if err != nil {
		return fmt.Errorf("Error deleting container %s in storage account %s : %v", containerName, storageAccountName, err)
	}

	// Now, we check and delete the storage account if its empty
	return nil
}

func getStorageAccountNameFromContainerUrl(containerUrl string) string {
	storageAccountName, _, _ := parseContainerUrl(containerUrl)
	return storageAccountName
}

func getContainerNameFromContainerUrl(containerUrl string) string {
	_, containerName, _ := parseContainerUrl(containerUrl)
	return containerName
}

func deleteAzureContainer(
	ctx context.Context,
	storageAccount,
	accessKey,
	containerName string) error {
	containerUrl, err := createContainerUrl(storageAccount, accessKey, containerName)

	if err != nil {
		return err
	}

	_, err = containerUrl.Delete(ctx, azblob.ContainerAccessConditions{})
	return err
}

func createContainerUrl(
	storageAccount string,
	accessKey string,
	containerName string) (azblob.ContainerURL, error) {
	// Create credentials
	credential, err := azblob.NewSharedKeyCredential(storageAccount, accessKey)
	if err != nil {
		return azblob.ContainerURL{}, fmt.Errorf("Invalid credentials with error : %v", err)
	}

	// Create a default request pipeline using credential
	pipeline := azblob.NewPipeline(credential, azblob.PipelineOptions{})

	urlString, err := url.Parse(fmt.Sprintf("https://%s.blob.core.windows.net", storageAccount))
	if err != nil {
		return azblob.ContainerURL{}, err
	}

	serviceURL := azblob.NewServiceURL(*urlString, pipeline)

	// Create containerURL that wraps the service url and pipeline to make requests
	containerURL := serviceURL.NewContainerURL(containerName)

	return containerURL, nil

}

func parseContainerUrl(containerUrl string) (string, string, string) {
	matches := storageAccountRE.FindStringSubmatch(containerUrl)
	storageAccount := matches[1]
	containerName := matches[2]
	blobName := matches[3]
	return storageAccount, containerName, blobName
}

func createAzureContainer(
	ctx context.Context,
	storageAccount string,
	accessKey string,
	containerName string,
	parameters map[string]string) (string, error) {
	if len(storageAccount) == 0 || len(accessKey) == 0 {
		return "", fmt.Errorf("Invalid storage account or access key")
	}

	containerURL, err := createContainerUrl(storageAccount, accessKey, containerName)
	if err != nil {
		return "", err
	}

	// Lets create a container with the containerURL
	_, err = containerURL.Create(ctx, parameters, azblob.PublicAccessNone)
	if err != nil {
		if serr, ok := err.(azblob.StorageError); ok {
			if serr.ServiceCode() == azblob.ServiceCodeBlobAlreadyExists {
				return containerURL.String(), nil
			}
		}
		return "", fmt.Errorf("Error creating container from containterURL : %s, Error : %v", containerURL.String(), err)
	}

	return containerURL.String(), nil
}
