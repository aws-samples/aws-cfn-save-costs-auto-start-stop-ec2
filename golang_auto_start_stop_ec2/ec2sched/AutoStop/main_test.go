package main

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/stretchr/testify/assert"
)

type MockEC2Client struct {
	ec2iface.EC2API
	DescribeInstancesOutputValue ec2.DescribeInstancesOutput
	StopInstancesOutputValue     ec2.StopInstancesOutput
}

type mockStopInst struct{}

func (m *MockEC2Client) DescribeInstances(input *ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error) {
	return &m.DescribeInstancesOutputValue, nil
}

func (m *MockEC2Client) StopInstances(input *ec2.StopInstancesInput) (*ec2.StopInstancesOutput, error) {
	return &m.StopInstancesOutputValue, nil
}

func (m mockStopInst) stopInst(instId string) (string, error) {
	return "i-0c938b5e573fb0f26", nil
}

// Output for mocked DescribeInstancesOutput from DesribeInstances EC2 api call
var outputZero = ec2.DescribeInstancesOutput{}
var outputWithAutostop = ec2.DescribeInstancesOutput{
	Reservations: []*ec2.Reservation{
		{
			Instances: []*ec2.Instance{
				{
					InstanceId:   aws.String("i-0c938b5e573fb0f26"),
					InstanceType: aws.String("m5.large"),
					State: &ec2.InstanceState{
						Name: aws.String(ec2.InstanceStateNameStopped),
					},
					Tags: []*ec2.Tag{
						{
							Key:   aws.String("autostart"),
							Value: aws.String("true"),
						},
						{
							Key:   aws.String("autostop"),
							Value: aws.String("true"),
						},
					},
				},
				{
					InstanceId:   aws.String("i-0c938b5e573fb0f27"),
					InstanceType: aws.String("m5.large"),
					State: &ec2.InstanceState{
						Name: aws.String(ec2.InstanceStateNameRunning),
					},
					Tags: []*ec2.Tag{
						{
							Key:   aws.String("autostop"),
							Value: aws.String("true"),
						},
						{
							Key:   aws.String("aws:autoscaling:groupName"),
							Value: aws.String("test-asg"),
						},
					},
				},
			},
		},
	},
}

func TestInstInfo(t *testing.T) {

	//Resp is the result from the ec2 api
	//Expected is the expected result from the function instInfo
	cases := []struct {
		Name     string
		Resp     ec2.DescribeInstancesOutput
		Expected ec2.DescribeInstancesOutput
	}{
		{
			Name:     "AutostopPresent",
			Resp:     outputWithAutostop,
			Expected: outputWithAutostop,
		},
		{
			Name:     "NoInstances",
			Resp:     outputZero,
			Expected: outputZero,
		},
	}

	//time to iterate over the test cases
	for _, c := range cases {

		//Sub test for each case
		t.Run(c.Name, func(t *testing.T) {
			e := ec2Api{
				Client: &MockEC2Client{
					EC2API:                       nil,
					DescribeInstancesOutputValue: c.Resp,
				},
			}
			inst, err := e.instInfo("autostop")

			if err != nil {
				fmt.Println("Unexpected Error - ", err.Error())
			}

			assert.Equal(t, "*ec2.DescribeInstancesOutput", fmt.Sprintf("%T", inst))

		})
	}
}

func TestEvalInst(t *testing.T) {

	//Resp is the result from the ec2 api StartInstances
	cases := []struct {
		Name          string
		InstanceInput *ec2.Instance
		Expected      string
	}{
		{
			Name:          "Autoscaling group present",
			InstanceInput: outputWithAutostop.Reservations[0].Instances[1],
			Expected:      "Skipping - i-0c938b5e573fb0f27 is part of autoscaling group test-asg",
		},
		{
			Name:          "Autoscaling group not present",
			InstanceInput: outputWithAutostop.Reservations[0].Instances[0],
			Expected:      "Stopping Instance: i-0c938b5e573fb0f26",
		},
	}

	//time to iterate over the test cases
	for _, c := range cases {

		//Sub test for each case
		t.Run(c.Name, func(t *testing.T) {
			s := mockStopInst{}
			st, err := evalInst(c.InstanceInput, s)

			if err != nil {
				fmt.Println("Unexpected Error - ", err.Error())
			}

			assert.Equal(t, c.Expected, st)

		})
	}

}

func TestStartInst(t *testing.T) {

	//Resp is the result from the ec2 api StartInstances
	cases := []struct {
		Name     string
		Resp     ec2.StopInstancesOutput
		Expected string
	}{
		{
			Name:     "StopInstancesOutput",
			Resp:     ec2.StopInstancesOutput{StoppingInstances: []*ec2.InstanceStateChange{{InstanceId: aws.String("i-0c938b5e573fb0f26")}}},
			Expected: "i-0c938b5e573fb0f26",
		},
	}

	//time to iterate over the test cases
	for _, c := range cases {

		//Sub test for each case
		t.Run(c.Name, func(t *testing.T) {
			e := ec2Api{
				Client: &MockEC2Client{
					EC2API:                   nil,
					StopInstancesOutputValue: c.Resp,
				},
			}
			inst, err := e.stopInst("i-0c938b5e573fb0f26")

			if err != nil {
				fmt.Println("Unexpected Error - ", err.Error())
			}

			assert.Equal(t, c.Expected, inst)

		})
	}
}
