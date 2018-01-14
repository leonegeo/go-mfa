// Copyright 2018, Radiant Solutions
// Licensed under the Apache License, Version 2.0

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/leonegeo/mfacache"
)

func main() {
	if len(os.Args) != 2 {
		usage()
	}

	switch os.Args[1] {

	case "set":
		doDelete()
		doSet()
		doShow()

	case "show":
		doShow()

	case "delete":
		doDelete()

	default:
		usage()
	}
}

func usage() {
	fmt.Printf("usage:\n")
	fmt.Printf("  $ mfa set      # removes your stored creds, then reads your MFA token and stores a new set\n")
	fmt.Printf("  $ mfa show     # reads your cached creds and displays them\n")
	fmt.Printf("  $ mfa delete   # removes your stored creds\n")

	os.Exit(1)
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func doSet() {
	err := mfacache.StoreCredentials(mfacache.DefaultDuration)
	check(err)
}

func doShow() {
	path, err := mfacache.GetCachePath(mfacache.DefaultProfile)
	check(err)

	byts, err := ioutil.ReadFile(path)
	check(err)

	value := map[string]interface{}{}
	err = json.Unmarshal(byts, &value)
	check(err)

	t, err := time.Parse(time.RFC3339, value["Expiration"].(string))
	check(err)
	d := t.Sub(time.Now()).Round(time.Second)

	fmt.Printf("Cache           %s\n", path)
	fmt.Printf("AccessKeyId     %s\n", value["AccessKeyID"])
	fmt.Printf("SecretAccessKey %s...\n", value["SecretAccessKey"].(string)[:4])
	fmt.Printf("SessionToken    %s...\n", value["SessionToken"].(string)[:4])
	fmt.Printf("ProviderName    %s\n", value["ProviderName"])
	fmt.Printf("Expiration      %s (%s)\n", value["Expiration"], d)

	sess, err := mfacache.NewSession()
	check(err)

	svc := iam.New(sess)
	input := iam.ListUsersInput{}
	output, err := svc.ListUsers(&input)
	check(err)
	fmt.Printf("User count:     %d\n", len(output.Users))
}

func doDelete() {
	path, err := mfacache.GetCachePath(mfacache.DefaultProfile)
	check(err)

	_ = os.Remove(path)
}
