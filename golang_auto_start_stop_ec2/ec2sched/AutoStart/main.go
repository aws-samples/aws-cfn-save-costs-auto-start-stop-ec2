package main

import (
	"fmt"
	"go/ec2sched/pkg/sess"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
)

type startInst interface {
	startInst(instId string) (string, error)
}

type instInfo interface {
	instInfo(tagName string) (*ec2.DescribeInstancesOutput, error)
}

type ec2Api struct {
	Client ec2iface.EC2API
}

// Fetch instances with tag "AutoSart", which is passed as input parameter
func (e ec2Api) instInfo(tagName string) (*ec2.DescribeInstancesOutput, error) {

	var maxOutput int = 75
	m := int64(maxOutput)
	var resp *ec2.DescribeInstancesOutput

	input := &ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("tag-key"),
				Values: []*string{
					aws.String(tagName),
				},
			},
		},
		MaxResults: &m,
	}

	//Cycle through paginated results for describe instances (incase we have more than 75 instances)
	for {
		instOutput, err := e.Client.DescribeInstances(input)
		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok {
				return nil, awsErr
			}
			return nil, err
		}

		if resp == nil {
			resp = instOutput
		} else {
			resp.Reservations = append(resp.Reservations, instOutput.Reservations...)
		}

		if instOutput.NextToken == nil {
			break
		}

		input.NextToken = instOutput.NextToken
	}
	//fmt.Println(resp)
	return resp, nil

}

// Evaluate instances to see if we can start.
func evalInst(inst *ec2.Instance, s startInst) (string, error) {

	fmt.Println(*inst.InstanceId)

	for _, tag := range inst.Tags {
		//ASGs not supported
		if *tag.Key == "aws:autoscaling:groupName" {
			return fmt.Sprintf("Skipping - %s is part of autoscaling group %s", string(*inst.InstanceId), *tag.Value), nil
		}
	}

	result, err := s.startInst(*inst.InstanceId)
	if err != nil {
		return "", err
	}
	return "Starting Instance: " + result, nil
}

func (e ec2Api) startInst(instId string) (string, error) {

	input := &ec2.StartInstancesInput{
		InstanceIds: []*string{
			aws.String(instId),
		},
	}

	res, si_err := e.Client.StartInstances(input)
	if si_err != nil {
		if awsErr, ok := si_err.(awserr.Error); ok {
			fmt.Println(awsErr.Error())
			return "", awsErr
		} else {
			fmt.Println(si_err.Error())
			return "", si_err
		}
	}
	return *res.StartingInstances[0].InstanceId, nil
}

func HandleLambdaEvent() {

	//Tag to look for on the instances
	schedule_tag := "autostart"
	region := os.Getenv("REGION_TZ")
	sess, err := sess.EstablishSession(region)
	e := ec2Api{
		Client: ec2.New(sess),
	}

	instDesc, err := e.instInfo(schedule_tag)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	instCount := 0

	for idx, res := range instDesc.Reservations {
		instCount += len(res.Instances)

		for _, inst := range instDesc.Reservations[idx].Instances {
			if *inst.State.Name != "running" {
				st, err := evalInst(inst, e)
				if err != nil {
					fmt.Println(err.Error())
				}
				fmt.Println(st)

			} else {
				fmt.Printf("Instance: %s is already running\n", *inst.InstanceId)
			}
		}
		fmt.Println("-----")
	}
	fmt.Printf("Instance count evaluated with %s tag: %d", schedule_tag, instCount)
	return
}

func main() {
	lambda.Start(HandleLambdaEvent)
}
