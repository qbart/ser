package main

import (
	"fmt"

	"github.com/go-ini/ini"
	// "github.com/aws/aws-sdk-go/aws/session"
)

func main() {

	cfg, err := ini.Load(
		"~/.aws/credentials",
	)
	if err != nil {
		fmt.Println(cfg)
		fmt.Println(cfg.Section(profile).Key("aws_access_key_id").String())
		fmt.Println(cfg.Section(profile).Key("aws_secret_access_key").String())
	} else {
		fmt.Println(err)
	}
	// sess := session.Must(session.NewSession())
}
