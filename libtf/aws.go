package libtf

import (
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
)

type ecsClient struct {
	c       *ecs.ECS
	cluster *string
}

func newClient(vault Vault, logHttp bool) *ecsClient {
	region := vault.AwsRegion()
	envName := vault.EnvName()
	sess, err := session.NewSession(&aws.Config{
		Region:      &region,
		Credentials: credentials.NewStaticCredentials(vault.AwsKey(), vault.AwsSecret(), ""),
	})
	if err != nil {
		panic(err)
	}
	config := &aws.Config{}
	if logHttp {
		config.LogLevel = aws.LogLevel(aws.LogDebugWithHTTPBody)
	}
	return &ecsClient{
		c:       ecs.New(sess, config),
		cluster: &envName,
	}
}

func (client *ecsClient) listInstances() ([]string, error) {
	out, err := client.c.ListContainerInstances(&ecs.ListContainerInstancesInput{
		Cluster: client.cluster,
	})
	if err != nil {
		return nil, err
	}
	res := []string{}
	for _, item := range out.ContainerInstanceArns {
		if item != nil {
			res = append(res, *item)
		}
	}
	return res, nil
}

func errorFromFailures(failures []*ecs.Failure) error {
	if len(failures) > 0 {
		reason := "unknown reason"
		for _, failure := range failures {
			if failure != nil {
				if failure.Reason != nil {
					reason = *failure.Reason
					break
				}
			}
		}
		return errors.New(reason)
	}
	return nil
}

func (client *ecsClient) runTask(instances []string, task string) error {
	containerInstances := []*string{}
	task = fmt.Sprintf("%s-%s", *client.cluster, task)
	for _, arn := range instances {
		containerInstances = append(containerInstances, &arn)
	}

	out, err := client.c.StartTask(&ecs.StartTaskInput{
		Cluster:            client.cluster,
		ContainerInstances: containerInstances,
		TaskDefinition:     &task,
	})

	if err != nil {
		return err
	}

	if err := errorFromFailures(out.Failures); err != nil {
		return err
	}

	tasks := []*string{}
	for _, task := range out.Tasks {
		if task != nil {
			if task.TaskArn != nil {
				tasks = append(tasks, task.TaskArn)
			}
		}
	}

	err = client.c.WaitUntilTasksStopped(&ecs.DescribeTasksInput{
		Cluster: client.cluster,
		Tasks:   tasks,
	})

	if err != nil {
		return err
	}

	describeResult, err := client.c.DescribeTasks(&ecs.DescribeTasksInput{
		Cluster: client.cluster,
		Tasks:   tasks,
	})

	if err != nil {
		return err
	}

	if err := errorFromFailures(describeResult.Failures); err != nil {
		return err
	}

	return nil
}

func RunEcsTask(vault Vault, task string, allInstances bool) error {
	client := newClient(vault, false)

	instances, err := client.listInstances()
	if err != nil {
		return err
	}

	runInstances := instances
	if !allInstances {
		runInstances = instances[:1]
	}

	fmt.Printf("running %s on %s\n", task, runInstances)

	if err := client.runTask(runInstances, task); err != nil {
		return err
	}

	return nil
}
