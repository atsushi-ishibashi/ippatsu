package svc

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/batch"
)

type BatchClient struct {
	*batch.Batch
}

func (bc *BatchClient) ListActiveJobDefinitions() (map[string][]int64, error) {
	jdefs := map[string][]int64{}
	input := &batch.DescribeJobDefinitionsInput{
		Status: aws.String("ACTIVE"),
	}
	result, err := bc.DescribeJobDefinitions(input)
	if err != nil {
		return jdefs, err
	}
	for _, v := range result.JobDefinitions {
		if jd, ok := jdefs[*v.JobDefinitionName]; ok {
			jdefs[*v.JobDefinitionName] = append(jd, *v.Revision)
		} else {
			jdefs[*v.JobDefinitionName] = []int64{*v.Revision}
		}
	}
	return jdefs, nil
}

func (bc *BatchClient) DescribeJobDefinition(name string) (*batch.JobDefinition, error) {
	strs := strings.Split(name, ":")
	if len(strs) != 2 {
		return nil, fmt.Errorf("Invalid JobDefinition name %s, expected name:revision", name)
	}
	input := &batch.DescribeJobDefinitionsInput{
		JobDefinitionName: aws.String(strs[0]),
	}
	result, err := bc.DescribeJobDefinitions(input)
	if err != nil {
		return nil, err
	}
	if len(result.JobDefinitions) < 1 {
		return nil, fmt.Errorf("Not Found Job Definition %s", name)
	}
	var jobDef *batch.JobDefinition
	for _, v := range result.JobDefinitions {
		if strconv.FormatInt(*v.Revision, 10) == strs[1] {
			jobDef = v
		}
	}
	return jobDef, nil
}

func (bc *BatchClient) DescribeJobQueue(name string) (*batch.JobQueueDetail, error) {
	input := &batch.DescribeJobQueuesInput{
		JobQueues: []*string{
			aws.String(name),
		},
	}
	result, err := bc.DescribeJobQueues(input)
	if err != nil {
		return nil, err
	}
	if len(result.JobQueues) == 0 {
		return nil, fmt.Errorf("Not Founf Job Queue %s", name)
	}
	queue := result.JobQueues[0]
	return queue, nil
}

func (bc *BatchClient) DescribeJobStatus(jobID string) (string, error) {
	input := &batch.DescribeJobsInput{
		Jobs: []*string{
			aws.String(jobID),
		},
	}
	result, err := bc.DescribeJobs(input)
	if err != nil {
		return "", err
	}
	job := result.Jobs[0]
	return *job.Status, nil
}

type JobQueue struct {
	name     string
	priority int64
	state    string
	status   string
}

func (jq JobQueue) String() string {
	return fmt.Sprintf("%s  Priority: %d, State: %s, Status: %s", jq.name, jq.priority, jq.state, jq.status)
}

func (bc *BatchClient) ListJobQueues() ([]JobQueue, error) {
	queues := []JobQueue{}
	input := &batch.DescribeJobQueuesInput{}
	result, err := bc.DescribeJobQueues(input)
	if err != nil {
		return queues, err
	}
	for _, v := range result.JobQueues {
		jq := JobQueue{
			name:     *v.JobQueueName,
			priority: *v.Priority,
			state:    *v.State,
			status:   *v.Status,
		}
		queues = append(queues, jq)
	}
	return queues, nil
}

type SubmitJob struct {
	ID   string
	Name string
}

func (bc *BatchClient) SubmitJobWithParams(jobname, jobdef, queue string) (*SubmitJob, error) {
	input := &batch.SubmitJobInput{
		JobDefinition: aws.String(jobdef),
		JobName:       aws.String(jobname),
		JobQueue:      aws.String(queue),
	}
	result, err := bc.SubmitJob(input)
	if err != nil {
		return nil, err
	}
	return &SubmitJob{ID: *result.JobId, Name: *result.JobName}, nil
}
