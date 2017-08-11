# ippatsu
ippatsu is CLI for AWS Batch.
## Usage
### submit
submit job and wait until job finish
```
$ ippatsu submit --help
NAME:
   ippatsu submit - submit job

USAGE:
   ippatsu submit [command options] [arguments...]

OPTIONS:
   --name value      job name. required
   --queue value     job queue name. required
   --jobdef value    job definition name. required
   --maxcheck value  max count to check job status (default: 60)
   --maxfail value   max count to miss checking job status in a row (default: 5)
   --dry-run         dry-run. print something without execute

EXAMPLES:
  ippatsu --awsconf hoge submit --name test --queue first-run-job-queue --jobdef first-run-job-definition:1 --dry-run
  ippatsu --awsconf hoge submit --name test --queue first-run-job-queue --jobdef first-run-job-definition:1
```
### list
```
$ ippatsu list --help
NAME:
   ippatsu list - list

USAGE:
   ippatsu list command [command options] [arguments...]

COMMANDS:
     jobdefs  list job definitions in active status
     queues   list job queues with some infos
```
**JobQueues**
```
$ ippatsu list queues
Job Queues:
  first-run-job-queue  Priority: 1, State: ENABLED, Status: VALID
```
**JobDefinitions**
```
$ ippatsu list jobdefs   
Job Definitions in active:
	first-run-job-definition (1, 2)
```
