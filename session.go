// Copyright 2018, Radiant Solutions
// Licensed under the Apache License, Version 2.0

package mfacache

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	ini "gopkg.in/ini.v1"
)

// CachedSession is the entire file
type CachedSession struct {
	Credentials      *CachedCredentials      `json:"Credentials"`
	AssumedRoleUser  *map[string]interface{} `json:"AssumedRoleUser"`
	ResponseMetadata *map[string]interface{} `json:"ResponseMetadata"`
}

// CachedCredentials represents the creds, plus it's expiration time
type CachedCredentials struct {
	AccessKeyID     string    `json:"AccessKeyId"`
	SecretAccessKey string    `json:"SecretAccessKey"`
	SessionToken    string    `json:"SessionToken"`
	Expiration      time.Time `json:"Expiration"`
}

// NewSession makes a session for hands-free usage: if a token
// is required, it'll just error out. You can use the supplied app to generate
// the token from the command line.
func NewSession() (*session.Session, error) {

	profile, ok := GetProfileName()
	if !ok {
		return session.NewSession()
	}

	cachedCreds, err := ReadCachedCredentials(profile)
	if err != nil {
		return nil, err
	}
	if cachedCreds == nil {
		return nil, errors.New("cached creds read failed: " + profile)
	}

	// static provider creds are never actually expired on the client side,
	// so we shall do an explicit check here ourselves
	if time.Now().After(cachedCreds.Expiration) {
		return nil, fmt.Errorf("MFA token has expired")
	}

	// make a set of official creds
	formalCreds := credentials.Value{
		AccessKeyID:     cachedCreds.AccessKeyID,
		SecretAccessKey: cachedCreds.SecretAccessKey,
		SessionToken:    cachedCreds.SessionToken,
		//ProviderName:    cachedCreds.ProviderName,
	}
	creds := credentials.NewStaticCredentialsFromCreds(formalCreds)

	// and make a proper session out of it: the Config settings, with the
	// creds, should override the normal AWS config files
	opts := session.Options{
		Config: aws.Config{
			Credentials: creds,
		},
		SharedConfigState: session.SharedConfigEnable,
		Profile:           profile,
	}

	sess, err := session.NewSessionWithOptions(opts)
	if err != nil {
		return nil, err
	}

	// verify the creds are working. Better to fail now than at some
	// random point later on.
	_, err = sess.Config.Credentials.Get()
	if err != nil {
		return nil, err
	}

	return sess, nil
}

// GetCachePath reads the config INI file to determine the name of the cache file
func GetCachePath(profile string) (string, error) {

	home, ok := os.LookupEnv("HOME")
	if !ok || home == "" {
		return "", fmt.Errorf("unable to read $HOME")
	}

	var path = home + "/.aws/config"

	cfg, err := ini.Load(path)
	if err != nil {
		return "", err
	}
	sec, err := cfg.GetSection("profile " + profile)
	if err != nil {
		return "", err
	}
	key, err := sec.GetKey("role_arn")
	if err != nil {
		return "", err
	}
	value := key.String()

	value = strings.Replace(value, ":", "_", -1)
	value = strings.Replace(value, "/", "-", -1)

	value = home + "/.aws/cli/cache/" + profile + "--" + value + ".json"

	return value, nil
}

// ReadCachedCredentials reads the cached creds
func ReadCachedCredentials(profile string) (*CachedCredentials, error) {
	path, err := GetCachePath(profile)
	if err != nil {
		return nil, err
	}

	_, err = os.Stat(path)
	if err != nil {
		return nil, errors.New("cache file not found: " + profile)
	}

	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	sessionData := &CachedSession{}

	err = json.Unmarshal(content, sessionData)
	if err != nil {
		return nil, err
	}

	return sessionData.Credentials, nil
}

// GetProfileName returns the profile name, or false if AWS_PROFILE is not set
func GetProfileName() (string, bool) {
	profile, ok := os.LookupEnv("AWS_PROFILE")
	if !ok {
		return "", false
	}
	if profile == "" {
		profile = session.DefaultSharedConfigProfile
	}
	return profile, true
}
