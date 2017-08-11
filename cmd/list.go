package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/atsushi-ishibashi/ippatsu/svc"
	"github.com/atsushi-ishibashi/ippatsu/util"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/batch"
	"github.com/urfave/cli"
)

func NewListCommand() cli.Command {
	return cli.Command{
		Name:  "list",
		Usage: "list",
		Subcommands: []cli.Command{
			newListJobDefsCommand(),
			newListJobQueuesCommand(),
		},
	}
}

func newListJobDefsCommand() cli.Command {
	return cli.Command{
		Name:  "jobdefs",
		Usage: "list job definitions in active status",
		Flags: []cli.Flag{},
		Action: func(c *cli.Context) error {
			if err := util.ConfigAWS(c); err != nil {
				return err
			}
			ls, err := newList(c)
			if err != nil {
				return err
			}
			return ls.printJobDefinitions()
		},
	}
}

func newListJobQueuesCommand() cli.Command {
	return cli.Command{
		Name:  "queues",
		Usage: "list job queues with some infos",
		Flags: []cli.Flag{},
		Action: func(c *cli.Context) error {
			if err := util.ConfigAWS(c); err != nil {
				return err
			}
			ls, err := newList(c)
			if err != nil {
				return err
			}
			return ls.printJobQueues()
		},
	}
}

type list struct {
	batchCli *svc.BatchClient
}

func newList(c *cli.Context) (*list, error) {
	ls := &list{}
	awsregion := os.Getenv("AWS_DEFAULT_REGION")
	sess, err := session.NewSession()
	if err != nil {
		return nil, err
	}
	ls.batchCli = &svc.BatchClient{Batch: batch.New(sess, aws.NewConfig().WithRegion(awsregion))}
	return ls, nil
}

func (ls *list) printJobDefinitions() error {
	jdefs, err := ls.batchCli.ListActiveJobDefinitions()
	if err != nil {
		return err
	}
	fmt.Println("Job Definitions in active:")
	for k, v := range jdefs {
		revStr := fmt.Sprintf("%d", v)
		revStr = strings.Replace(revStr, "[", "", -1)
		revStr = strings.Replace(revStr, "]", "", -1)
		revStr = strings.Replace(revStr, " ", ", ", -1)
		fmt.Println("\t" + k + " (" + revStr + ")")
	}
	return nil
}

func (ls *list) printJobQueues() error {
	jqs, err := ls.batchCli.ListJobQueues()
	if err != nil {
		return err
	}
	fmt.Println("Job Queues:")
	for _, v := range jqs {
		fmt.Println("\t" + v.String())
	}
	return nil
}
