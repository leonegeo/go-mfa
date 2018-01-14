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
)

// CachedCredential represents the creds, plus it's expiration time
type CachedCredential struct {
	credentials.Value
	Expiration time.Time
}

// GetCachePath returns location of the cache file
func GetCachePath(profile string) (string, error) {
	home := os.Getenv("HOME")
	if home == "" {
		return "", fmt.Errorf("unable to read $HOME")
	}

	path := filepath.Join(home, DefaultLocation)

	file := fmt.Sprintf("mfacache-%s.json", profile)
	path = filepath.Join(path, file)

	return path, nil
}

func (c *CachedCredential) Read(profile string) error {
	path, err := GetCachePath(profile)
	if err != nil {
		return err
	}

	_, err = os.Stat(path)
	if err != nil {
		return err
	}

	_, err = os.Stat(path)
	if err != nil {
		return err
	}

	content, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	err = json.Unmarshal(content, c)
	if err != nil {
		return err
	}

	return nil
}

func (c *CachedCredential) Write(profile string) error {

	path, err := GetCachePath(profile)
	if err != nil {
		return err
	}

	byts, err := json.Marshal(c)
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, 0700)
	}

	err = ioutil.WriteFile(path, byts, 0600)
	if err != nil {
		return err
	}

	return nil
}
