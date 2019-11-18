package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/go-ini/ini"
)

var (
	credsFile = os.Getenv("HOME") + "/.aws/credentials"
	cfgFile   = os.Getenv("HOME") + "/.aws/config"
)

func fatalErr(err error, msg string) {
	if err != nil {
		log.Fatal(msg+":", err)
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go-aws-mfa [profile]")
		os.Exit(1)
	}

	profile := os.Args[1]
	fmt.Printf("Authenticating for %s\n", profile)

	credsini, err := ini.Load(credsFile)
	fatalErr(err, "Failed to load credentials file")

	srcProfile := credsini.Section(profile).Key("long_term").MustString("")
	assumeRole := credsini.Section(profile).Key("assume_role").MustString("")
	if srcProfile == "" {
		srcProfile = profile
	}

	ltProfile := srcProfile + "-long-term"
	cred := credsini.Section(ltProfile)
	mfasn := cred.Key("aws_mfa_device").MustString("")
	id := cred.Key("aws_access_key_id").MustString("")
	key := cred.Key("aws_secret_access_key").MustString("")

	fmt.Println("Sourcing creds from", ltProfile)
	if id == "" || key == "" {
		fmt.Println("ERROR: couldn't find key id or access key")
		os.Exit(1)
	}

	if assumeRole != "" {
		fmt.Println("Assuming role", assumeRole)
	}

	fmt.Println("Using the MFA device", mfasn)

	sess, err := session.NewSession(&aws.Config{Credentials: credentials.NewSharedCredentials("", ltProfile)})
	fatalErr(err, "ERROR: failed to create session")

	fmt.Printf("Enter MFA code: ")
	r := bufio.NewReader(os.Stdin)
	code, _, err := r.ReadLine()
	fatalErr(err, "ERROR: failed to read MFA code")
	mfaCode := string(code)

	if len(code) != 6 {
		fmt.Println("ERROR: code must be 6 digits")
		os.Exit(1)
	}

	stSec, err := credsini.NewSection(profile)
	fatalErr(err, "ERROR: failed to create a new section for "+profile)

	validUntil := time.Now()
	_sts := sts.New(sess)
	if assumeRole == "" {
		res, err := _sts.GetSessionToken(&sts.GetSessionTokenInput{
			TokenCode:    &mfaCode,
			SerialNumber: &mfasn,
		})
		fatalErr(err, "ERROR: failed to get session token")

		stSec.NewKey("aws_access_key_id", *res.Credentials.AccessKeyId)
		stSec.NewKey("aws_secret_access_key", *res.Credentials.SecretAccessKey)
		stSec.NewKey("aws_session_token", *res.Credentials.SessionToken)
		validUntil = *res.Credentials.Expiration

	} else {
		res, err := _sts.AssumeRole(&sts.AssumeRoleInput{
			TokenCode:       &mfaCode,
			SerialNumber:    &mfasn,
			RoleArn:         &assumeRole,
			RoleSessionName: &profile,
		})
		fatalErr(err, "ERROR: failed to assume role")

		stSec.NewKey("aws_access_key_id", *res.Credentials.AccessKeyId)
		stSec.NewKey("aws_secret_access_key", *res.Credentials.SecretAccessKey)
		stSec.NewKey("aws_session_token", *res.Credentials.SessionToken)
		validUntil = *res.Credentials.Expiration
	}

	err = credsini.SaveTo(credsFile)
	fatalErr(err, "ERROR: failed to update profile with new credentials")

	fmt.Printf("Credentials updated for %s, valid until %s\n", profile, validUntil.In(time.Local))
}
