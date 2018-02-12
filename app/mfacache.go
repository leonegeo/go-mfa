// Copyright 2018, Radiant Solutions
// Licensed under the Apache License, Version 2.0

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"time"

	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/leonegeo/mfacache"
)

func main() {

	switch os.Args[1] {

	case "show":
		doShow()

	case "set":
		doSet()

	case "delete":
		doDelete()

	default:
		usage()
	}
}

func usage() {
	fmt.Printf("usage:\n")
	fmt.Printf("  $ mfa [--profile NAME] show     # reads your cached creds and displays them\n")
	fmt.Printf("  $ mfa [--profile NAME] delete   # removes your stored creds\n")

	os.Exit(1)
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func doSet() {
	cmd := exec.Command("aws", "s3", "ls")
	err := cmd.Run()
	check(err)
}

func doShow() {
	profile, ok := os.LookupEnv("AWS_PROFILE")
	if !ok || profile == "" {
		profile = "default"
	}

	path, err := mfacache.GetCachePath(profile)
	check(err)

	_, err = os.Stat(path)
	if err != nil {
		check(errors.New("cache file not found: " + profile))
	}

	byts, err := ioutil.ReadFile(path)
	check(err)

	value := map[string]interface{}{}
	err = json.Unmarshal(byts, &value)
	check(err)

	creds, ok := value["Credentials"].(map[string]interface{})
	if !ok {
		check(errors.New("unable to parse creds"))
	}

	t, err := time.Parse(time.RFC3339, creds["Expiration"].(string))
	check(err)
	d := t.Sub(time.Now()).Round(time.Second)

	fmt.Printf("Profile         %s\n", profile)
	fmt.Printf("Cache           %s\n", path)
	fmt.Printf("AccessKeyId     %s\n", creds["AccessKeyId"].(string))
	fmt.Printf("SecretAccessKey %s...\n", creds["SecretAccessKey"].(string)[:4])
	fmt.Printf("SessionToken    %s...\n", creds["SessionToken"].(string)[:4])
	fmt.Printf("Expiration      %s (%s)\n", creds["Expiration"], d)

	sess, err := mfacache.NewSession()
	check(err)

	svc := iam.New(sess)
	input := iam.ListGroupsInput{}
	output, err := svc.ListGroups(&input)
	check(err)
	fmt.Printf("Groups count:     %d\n", len(output.Groups))
}

func doDelete() {
	profile, ok := os.LookupEnv("AWS_PROFILE")
	if !ok || profile == "" {
		profile = "default"
	}

	path, err := mfacache.GetCachePath(profile)
	check(err)

	_, err = os.Stat(path)
	if err != nil {
		check(errors.New("cache file not found: " + profile))
	}

	_ = os.Remove(path)
}
