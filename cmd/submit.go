package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/atsushi-ishibashi/ippatsu/svc"
	"github.com/atsushi-ishibashi/ippatsu/util"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/batch"
	"github.com/urfave/cli"
)

func NewSubmitCommand() cli.Command {
	return cli.Command{
		Name:  "submit",
		Usage: "submit job",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "name",
				Usage: "job name. required",
			},
			cli.StringFlag{
				Name:  "queue",
				Usage: "job queue name. required",
			},
			cli.StringFlag{
				Name:  "jobdef",
				Usage: "job definition name. required",
			},
			cli.IntFlag{
				Name:  "maxcheck",
				Usage: "max count to check job status",
				Value: maxCheckCount,
			},
			cli.IntFlag{
				Name:  "maxfail",
				Usage: "max count to miss checking job status in a row",
				Value: maxFailCount,
			},
			cli.BoolFlag{
				Name:  "dry-run",
				Usage: "dry-run. print something without execute",
			},
		},
		Action: func(c *cli.Context) error {
			printErr := func(e error) error {
				return util.ErrorRed(e.Error())
			}
			if err := util.ConfigAWS(c); err != nil {
				return printErr(err)
			}
			sm, err := newSubmit(c)
			if err != nil {
				return printErr(err)
			}
			if c.Int("maxfail") != 0 {
				maxFailCount = c.Int("maxfail")
			}
			if c.Int("maxcheck") != 0 {
				maxCheckCount = c.Int("maxcheck")
			}
			if c.Bool("dry-run") {
				if err = sm.dryPrint(); err != nil {
					return printErr(err)
				}
			} else {
				if err = sm.dryPrint(); err != nil {
					return printErr(err)
				}
				err = sm.execute()
				if err != nil {
					return printErr(err)
				}
			}
			return nil
		},
	}
}

type submit struct {
	batchCli *svc.BatchClient
	jobName  string
	queue    string
	jobDef   string
}

func newSubmit(c *cli.Context) (*submit, error) {
	sm := &submit{}

	if c.String("name") == "" {
		return nil, fmt.Errorf("--name is required")
	}
	sm.jobName = c.String("name")

	if c.String("queue") == "" {
		return nil, fmt.Errorf("--queue is required")
	}
	sm.queue = c.String("queue")

	if c.String("jobdef") == "" {
		return nil, fmt.Errorf("--jobdef is required")
	}
	sm.jobDef = c.String("jobdef")

	awsregion := os.Getenv("AWS_DEFAULT_REGION")
	sess, err := session.NewSession()
	if err != nil {
		return nil, err
	}
	sm.batchCli = &svc.BatchClient{Batch: batch.New(sess, aws.NewConfig().WithRegion(awsregion))}
	return sm, nil
}

func (sm *submit) String() string {
	return "Job to submit:\n" + "\tname: " + sm.jobName + "\n" + "\tjob definition: " + sm.jobDef + "\n" + "\tjob queue: " + sm.queue
}

func (sm *submit) execute() error {
	jb, err := sm.batchCli.SubmitJobWithParams(sm.jobName, sm.jobDef, sm.queue)
	if err != nil {
		return err
	}
	util.PrintlnGreen("Submitted JobName: " + jb.Name + ", ID: " + jb.ID)
	return sm.waitUntilJobFinish(jb.ID, jb.Name)
}

func (sm *submit) dryPrint() error {
	fmt.Println(sm)
	jobdef, err := sm.batchCli.DescribeJobDefinition(sm.jobDef)
	if err != nil {
		return err
	}
	fmt.Println("Job Definition Details:")
	fmt.Println(jobdef.String())
	jobque, err := sm.batchCli.DescribeJobQueue(sm.queue)
	if err != nil {
		return err
	}
	fmt.Println("Job Queue Details:")
	fmt.Println(jobque.String())
	return nil
}

type jobStatus struct {
	id, name, status string
}

var (
	maxFailCount  = 5
	maxCheckCount = 60
)

func (sm *submit) waitUntilJobFinish(jobID, name string) error {
	jobSt := jobStatus{id: jobID, name: name, status: batch.JobStatusSubmitted}
	failCount := maxFailCount
	remainingCheckCount := maxCheckCount
	for remainingCheckCount > 0 {
		currentStatus, err := sm.batchCli.DescribeJobStatus(jobID)
		if err != nil {
			failCount--
			if failCount < 1 {
				return fmt.Errorf("Fail to fetch job status %d times JobID: %s current status %s", maxFailCount, jobID, jobSt.status)
			}
			continue
		}
		if currentStatus != jobSt.status {
			switch currentStatus {
			case batch.JobStatusPending,
				batch.JobStatusRunnable,
				batch.JobStatusStarting,
				batch.JobStatusRunning:
				util.PrintlnGreen(fmt.Sprintf("JobID: %s change status %s -> %s", jobID, jobSt.status, currentStatus))
			case batch.JobStatusSucceeded:
				util.PrintlnGreen(fmt.Sprintf("JobID: %s change status %s -> %s", jobID, jobSt.status, currentStatus))
			case batch.JobStatusFailed:
				util.PrintlnRed(fmt.Sprintf("JobID: %s change status %s -> %s", jobID, jobSt.status, currentStatus))
			}
			jobSt.status = currentStatus
		}
		if jobSt.status == batch.JobStatusSucceeded {
			util.PrintlnGreen(fmt.Sprintf("JobName: %s, JobID: %s SUCEEDED", name, jobID))
			return nil
		}
		if jobSt.status == batch.JobStatusFailed {
			return fmt.Errorf("JobName: %s, JobID: %s FAILED", name, jobID)
		}
		time.Sleep(10 * time.Second)
		failCount = maxFailCount
		remainingCheckCount--
	}
	if remainingCheckCount < 1 {
		return fmt.Errorf("Consumed maxCheckCount(%d) JobID: %s current status %s", maxCheckCount, jobID, jobSt.status)
	}
	return nil
}
