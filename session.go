// Copyright 2018, Radiant Solutions
// Licensed under the Apache License, Version 2.0

package mfacache

import (
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
)

// DefaultProfile is the, umm, default profile
const (
	DefaultProfile = "default"

	// DefaultDuration is how long until the creds expire
	DefaultDuration = time.Minute * time.Duration(60)

	// DefaultDir is where the creds cache file lives
	// (prefixed with $HOME)
	DefaultLocation = ".aws"
)

// StoreCredentials will ask for your token (via stdin) and then
// store the resulting credentials in the cache.
//
// AWS will error unless duration is the range [900...3600] seconds
func StoreCredentials(profile string, duration time.Duration) error {

	// there is really no simple way to change the token duration:
	// this trick actually works just fine for our purposes
	stscreds.DefaultDuration = duration

	// make a session that reads both(!) config files and knows to prompt
	// for our MFA token when needed
	opts := session.Options{
		SharedConfigState:       session.SharedConfigEnable,
		AssumeRoleTokenProvider: stscreds.StdinTokenProvider,
		Profile:                 profile,
	}
	sess, err := session.NewSessionWithOptions(opts)

	// force the MFA token to be read
	value, err := sess.Config.Credentials.Get()
	if err != nil {
		return err
	}

	// store the creds
	cachedCreds := &CachedCredential{value, time.Now().UTC().Add(duration)}
	err = cachedCreds.Write(profile)
	if err != nil {
		return err
	}

	return nil
}

// NewSession makes a session for hands-free usage: if a token
// is required, it'll just error out. You can use the supplied app to generate
// the token from the command line.
func NewSession(profile string) (*session.Session, error) {

	useMFA, ok := os.LookupEnv("MFACACHE")
	if !ok || useMFA != "1" {
		return session.NewSession()
	}

	// read the cached creds
	cachedCreds := &CachedCredential{}
	err := cachedCreds.Read(profile)
	if err != nil {
		return nil, err
	}

	// static provider creds are never actually expired on the client side,
	// so we shall do an explicit check here ourselves
	if time.Now().After(cachedCreds.Expiration) {
		return nil, fmt.Errorf("MFA token has expired")
	}

	// make a set of official creds
	creds := credentials.NewStaticCredentialsFromCreds(cachedCreds.Value)

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
