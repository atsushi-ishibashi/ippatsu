package main

import (
	"os"

	"github.com/atsushi-ishibashi/ippatsu/cmd"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "awsconf",
			Usage: "set env var from ~/.aws/credentials in process",
		},
		cli.StringFlag{
			Name:  "awsregion",
			Usage: "set AWS_DEFAULT_REGION in process",
			Value: "ap-northeast-1",
		},
	}

	submitCommand := cmd.NewSubmitCommand()
	listCommand := cmd.NewListCommand()

	app.Commands = []cli.Command{
		submitCommand,
		listCommand,
	}
	app.Run(os.Args)

}
