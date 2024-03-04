package main

import (
	"fmt"
	"go/ec2sched/pkg/sess"
	"go/ec2sched/pkg/settz"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
)

type stoptInst interface {
	stopInst(instId string, region string) (string, error)
}

type instInfo interface {
	instInfo(tagName string, region string) (*ec2.DescribeInstancesOutput, error)
}

type ec2Api struct {
	Client ec2iface.EC2API
}

func (e ec2Api) instInfo(tagName string, region string) (*ec2.DescribeInstancesOutput, error) {
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

func (e ec2Api) stopInst(instId string, region string) (string, error) {

	input := &ec2.StopInstancesInput{
		InstanceIds: []*string{
			aws.String(instId),
		},
	}

	res, si_err := e.Client.StopInstances(input)
	if si_err != nil {
		if awsErr, ok := si_err.(awserr.Error); ok {
			fmt.Println(awsErr.Error())
			return "", awsErr
		} else {
			fmt.Println(si_err.Error())
			return "", si_err
		}
	}
	//fmt.Println(res)
	return *res.StoppingInstances[0].InstanceId, nil
}

func evalInst(inst *ec2.Instance, region string, curTime time.Time, s stoptInst) (string, error) {
	dayOfWeek := curTime.Weekday()
	modTime := curTime.Format(("15:04"))

	//stopTime is the desired stop time for the ec2 instance
	stopTime := ""

	fmt.Println(*inst.InstanceId)

	//Check to see if weekend
	if int(dayOfWeek) >= 6 && int(dayOfWeek) <= 7 {
		return fmt.Sprintf("Current day of week %s. StopWeekDay requires non weekend values", dayOfWeek), nil
	}

	for _, tag := range inst.Tags {
		if *tag.Key == "StopWeekDay" {
			//fmt.Println("StopWeekDay tag found. Value: ", *tag.Value)
			stopTime = *tag.Value
		}
		//ASGs not supported
		if *tag.Key == "aws:autoscaling:groupName" {
			return fmt.Sprintf("Skipping - %s is part of autoscaling group %s", string(*inst.InstanceId), *tag.Value), nil
		}
	}

	//Point here is to get the times on the same day to compare time of day, not just the date. Ensuring Apples to Apples
	cur_tod, _ := time.Parse("15:04", modTime)
	stop_tod, stop_tod_err := time.Parse("15:04", stopTime)

	if stop_tod_err != nil {
		return "", stop_tod_err
	}

	cur_minus := cur_tod.Add(-time.Minute * 5)
	cur_plus := cur_tod.Add(time.Minute * 5)

	if stop_tod.After(cur_minus) && stop_tod.Before(cur_plus) {
		result, err := s.stopInst(*inst.InstanceId, region)
		if err != nil {
			return "", err
		}
		return "Stopping Instance: " + result, nil
	} else {
		strVal := string(*inst.InstanceId)
		return "StopWeekDay schedule not matched for: " + strVal, nil
	}
}

// Handler doing bulk of work
// Environment variables set at the lambda configuration level
func HandleLambdaEvent() {

	//Tag to look for on the instances
	schedule_tag := "stopweekday"
	tz, err := settz.SetRegion(os.Getenv("TZ"))
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(tz)
	fmt.Println(time.Now())

	region := os.Getenv("REGION_TZ")

	sess, err := sess.EstablishSession(region)
	e := ec2Api{
		Client: ec2.New(sess),
	}

	instDesc, err := e.instInfo(schedule_tag, region)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	instCount := 0

	for idx, res := range instDesc.Reservations {
		instCount += len(res.Instances)

		for _, inst := range instDesc.Reservations[idx].Instances {
			if *inst.State.Name != "stopped" {
				st, err := evalInst(inst, region, time.Now(), e)
				if err != nil {
					fmt.Println(err.Error())
				}
				fmt.Println(st)

			} else {
				fmt.Printf("Instance: %s is already stopped\n", *inst.InstanceId)
			}
		}
		fmt.Println("-----")
	}
	fmt.Printf("Instance count evaluated with %s tag: %d", schedule_tag, instCount)
	return
}

// Go must call main function first, so we call the handler from the main.
func main() {
	lambda.Start(HandleLambdaEvent)
}
