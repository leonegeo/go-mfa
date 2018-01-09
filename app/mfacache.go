// Copyright 2018, Radiant Solutions
// Licensed under the Apache License, Version 2.0package mfa

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/leonegeo/mfacache"
)

func main() {
	if len(os.Args) != 2 {
		usage()
	}

	switch os.Args[1] {

	case "set":
		doSet()

	case "show":
		doShow()

	case "delete":
		doDelete()

	case "reset":
		doDelete()
		doSet()

	default:
		usage()
	}
}

func usage() {
	fmt.Printf("usage:\n")
	fmt.Printf("  $ mfa set      # reads your MFA token and stores your creds\n")
	fmt.Printf("  $ mfa show     # reads your cached creds and displays them\n")
	fmt.Printf("  $ mfa delete   # removes your stored creds\n")
	fmt.Printf("  $ mfa reset    # same as 'mfa delete ; mfa set'\n")

	os.Exit(1)
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func doSet() {
	sess, err := mfacache.NewSession()
	check(err)

	// force the issue
	_, err = sess.Config.Credentials.Get()
	check(err)

	doShow()
}

func doShow() {
	path, file, err := mfacache.GetCachePath("")
	check(err)

	byts, err := ioutil.ReadFile(path + "/" + file)
	check(err)

	value := map[string]interface{}{}
	err = json.Unmarshal(byts, &value)
	check(err)

	t, err := time.Parse(time.RFC3339, value["Expiration"].(string))
	check(err)
	d := t.Sub(time.Now()).Round(time.Second)

	fmt.Printf("Cache           %s/%s\n", path, file)
	fmt.Printf("AccessKeyId     %s\n", value["AccessKeyID"])
	fmt.Printf("SecretAccessKey %s...\n", value["SecretAccessKey"].(string)[:4])
	fmt.Printf("SessionToken    %s...\n", value["SessionToken"].(string)[:4])
	fmt.Printf("ProviderName    %s\n", value["ProviderName"])
	fmt.Printf("Expiration      %s (%s)\n", value["Expiration"], d)
}

func doDelete() {
	path, file, err := mfacache.GetCachePath("")
	check(err)

	_ = os.Remove(path + "/" + file)
}
