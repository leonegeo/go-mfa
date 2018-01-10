// Copyright 2018, Radiant Solutions
// Licensed under the Apache License, Version 2.0

package mfacache

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
)

// NewSession returns a new session, using the cache: you can use this
// function as an example of how to do things in your own app.
// This is an interactive session, i.e. it will ask on stdin for your token.
func NewSession(duration time.Duration) (*session.Session, error) {

	opts := session.Options{
		SharedConfigState:       session.SharedConfigEnable,
		AssumeRoleTokenProvider: stscreds.StdinTokenProvider,
	}
	sess, err := session.NewSessionWithOptions(opts)
	if err != nil {
		return nil, err
	}

	provider := &FileCacheProvider{
		Creds:    sess.Config.Credentials,
		Duration: duration,
	}

	// Inject cache able credential provider on top of the SDK's credentials loader
	sess.Config.Credentials = credentials.NewCredentials(provider)

	_, err = sess.Config.Credentials.Get()
	if err != nil {
		return nil, err
	}
	return sess, nil
}

// NewNoninteractiveSession makes a session for hands-free usage: if a token
// is required, it'll just error out. You can use the supplied app to generate
// the token from the command line.
func NewNoninteractiveSession() (*session.Session, error) {

	f := func() (string, error) {
		return "", fmt.Errorf("MFA token not cached")
	}

	opts := session.Options{
		SharedConfigState:       session.SharedConfigEnable,
		AssumeRoleTokenProvider: f,
	}
	sess, err := session.NewSessionWithOptions(opts)
	if err != nil {
		return nil, err
	}

	provider := &FileCacheProvider{
		Creds:    sess.Config.Credentials,
		Duration: 0,
	}

	// Inject cache able credential provider on top of the SDK's credentials loader
	sess.Config.Credentials = credentials.NewCredentials(provider)

	_, err = sess.Config.Credentials.Get()
	if err != nil {
		return nil, err
	}
	return sess, nil
}
