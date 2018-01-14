// Copyright 2018, Radiant Solutions
// Licensed under the Apache License, Version 2.0

package mfacache

import (
	"fmt"
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
func StoreCredentials(duration time.Duration) error {

	// there is really no simple way to change the token duration:
	// this trick actually works just fine for our purposes
	stscreds.DefaultDuration = duration

	// make a session that reads both(!) config files and knows to prompt
	// for our MFA token when needed
	opts := session.Options{
		SharedConfigState:       session.SharedConfigEnable,
		AssumeRoleTokenProvider: stscreds.StdinTokenProvider,
	}
	sess, err := session.NewSessionWithOptions(opts)

	// force the MFA token to be read
	value, err := sess.Config.Credentials.Get()
	if err != nil {
		return err
	}

	// store the creds
	cachedCreds := &CachedCredential{value, time.Now().UTC().Add(duration)}
	err = cachedCreds.Write(DefaultProfile)
	if err != nil {
		return err
	}

	return nil
}

// NewSession makes a session for hands-free usage: if a token
// is required, it'll just error out. You can use the supplied app to generate
// the token from the command line.
func NewSession() (*session.Session, error) {

	// read the cached creds
	cachedCreds := &CachedCredential{}
	err := cachedCreds.Read(DefaultProfile)
	if err != nil {
		return nil, err
	}

	// static provider creds are never actually expire don the client side,
	// so we shall do an explicit check here ourselves
	if time.Now().After(cachedCreds.Expiration) {
		return nil, fmt.Errorf("MFA token has expired")
	}

	// make a set of official creds
	creds := credentials.NewStaticCredentialsFromCreds(cachedCreds.Value)

	// and make a proper session out of it
	config := aws.Config{
		Credentials: creds,
	}
	sess := session.New(&config)

	// verify the creds are working. Better to fail now than at some
	// random point later on.
	_, err = sess.Config.Credentials.Get()
	if err != nil {
		return nil, err
	}

	return sess, nil
}
