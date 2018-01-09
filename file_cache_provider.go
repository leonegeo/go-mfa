/*
Copyright 2017 WALLIX
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// fileCacheProvider was taken from https://github.com/wallix/awless, file
// credentials_providers.go. The only real change I made was to use
// a different cache file path.

package mfacache

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
)

// lives under $HOME
const credsDir = ".aws/mfacache"

//---------------------------------------------------------------------

type cachedCredential struct {
	credentials.Value
	Expiration time.Time
}

func (c *cachedCredential) isExpired() bool {
	return c.Expiration.Before(time.Now().UTC())
}

// FileCacheProvider implements the Provider interface to cache
// the creds in $HOME/.aws/mfacache/aws-profile-PROFILE.json.
type FileCacheProvider struct {
	Creds   *credentials.Credentials
	curr    *cachedCredential
	profile string
}

// GetCachePath returns location of the cache file
func GetCachePath(profile string) (string, string, error) {
	home := os.Getenv("HOME")
	if home == "" {
		return "", "", fmt.Errorf("unable to read $HOME")
	}

	credFolder := filepath.Join(home, credsDir)

	credFile := fmt.Sprintf("aws-profile-%s.json", profile)

	return credFolder, credFile, nil
}

// Retrieve returns the credentials, via a file cache.
func (f *FileCacheProvider) Retrieve() (credentials.Value, error) {
	credFolder, credFile, err := GetCachePath(f.profile)
	if err != nil {
		return credentials.Value{}, err
	}

	if content, ok := getFileContent(credFolder, credFile); ok {
		var cached *cachedCredential
		if err := json.Unmarshal(content, &cached); err != nil {
			return credentials.Value{}, err
		}
		//log.Printf("loading credentials from '%s'", filepath.Join(credFolder, credFile))
		if !cached.isExpired() {
			f.curr = cached
			return cached.Value, nil
		}
		f.Creds.Expire()

	}
	credValue, err := f.Creds.Get()
	if err != nil {
		/*if batcherr, ok := err.(awserr.BatchedErrors); !ok || batcherr.Code() != "NoCredentialProviders" {
			if failure, ok := err.(awserr.RequestFailure); ok {
				//log.Printf("%s: %s\n", failure.Code(), failure.Message())
			} else {
				//log.Printf("%s\n", err)
			}
		}*/
		return credValue, err
	}

	switch credValue.ProviderName {
	case stscreds.ProviderName:
		cred := &cachedCredential{credValue, time.Now().UTC().Add(stscreds.DefaultDuration)}
		f.curr = cred
		content, err := json.Marshal(cred)
		if err != nil {
			return credValue, err
		}
		if err = putFileContent(credFolder, credFile, content); err != nil {
			return credValue, fmt.Errorf("error writing cache file: %s", err.Error())
		}
		//log.Printf("credentials cached in '%s'", filepath.Join(credFolder, credFile))
		return credValue, nil
	}
	return credValue, nil
}

// IsExpired returns true iff the creds have expired.
func (f *FileCacheProvider) IsExpired() bool {
	if f.curr != nil {
		return f.curr.isExpired()
	}
	return f.Creds.IsExpired()
}

func getFileContent(path string, filename string) (content []byte, ok bool) {
	if _, err := os.Stat(path); err != nil {
		return
	}
	credPath := filepath.Join(path, filename)

	if _, readerr := os.Stat(credPath); readerr != nil {
		return
	}
	var err error
	if content, err = ioutil.ReadFile(credPath); err != nil {
		return
	}
	ok = true
	return
}

func putFileContent(path string, filename string, content []byte) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.MkdirAll(path, 0700)
	}

	return ioutil.WriteFile(filepath.Join(path, filename), content, 0600)
}
