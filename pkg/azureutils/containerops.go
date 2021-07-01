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
)

const (
	AccessKey = "GYpkP/vgSAH8vJYD/FrSysQOnTJ2nKsy1PFR2YSRWsoR9FMcc4PNIZXunAgrVpE/2Qbc86MfcDNhyKJqS7sUbA=="
)

var (
	storageAccountRE = regexp.MustCompile(`https://(.+).blob.core.windows.net/(.*)/(?:.*)`)
)

func CreateContainer(
	ctx context.Context,
	storageAccount string,
	accessKey string,
	containerName string,
	parameters map[string]string) (string, error) {
	if len(storageAccount) == 0 || len(accessKey) == 0 {
		return "", fmt.Errorf("Invalid storage account or access key")
	}

	containerURL, err := createContainerUrl(storageAccount, accessKey, containerName)
	if err ! = nil {
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

func DeleteContainer(
	ctx context.Context,
	containerURL string) error {
	matches := storageAccountRE.FindStringSubmatch(containerURL)
	storageAccount := matches[1]
	containerName := matches[2]
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

	blobServiceURL, err := url.Parse(
		fmt.Sprintf("https://%s.blob.core.windows.net/%s", storageAccount, containerName))
	if err != nil {
		return azblob.ContainerURL{}, fmt.Errorf("Unable to parse service url for blob storage")
	}

	// Create containerURL that wraps the service url and pipeline to make requests
	containerURL := azblob.NewContainerURL(*blobServiceURL, pipeline)
	return containerURL, nil

}
