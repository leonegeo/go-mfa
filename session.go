// Copyright 2018, Radiant Solutions
// Licensed under the Apache License, Version 2.0

package mfacache

import (
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
)

// NewSession returns a new session, using the cache: you can use this
// function as an example of how to do things in your own app.
func NewSession() (*session.Session, error) {

	opts := session.Options{
		SharedConfigState:       session.SharedConfigEnable,
		AssumeRoleTokenProvider: stscreds.StdinTokenProvider,
	}
	sess, err := session.NewSessionWithOptions(opts)
	if err != nil {
		return nil, err
	}

	// Inject cache able credential provider on top of the SDK's credentials loader
	sess.Config.Credentials = credentials.NewCredentials(&FileCacheProvider{
		Creds: sess.Config.Credentials,
	})

	return sess, nil
}
