package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/stretchr/testify/assert"
)

// Create a mock structure implementing the ec2iface interface. Makes the
// call to aws ec2 api
type MockEC2Client struct {
	ec2iface.EC2API
	DescribeInstancesOutputValue ec2.DescribeInstancesOutput
	StartInstancesOutputValue    ec2.StartInstancesOutput
}

// Implements the interface
type mockStartInst struct{}

// Implement the mocked DescribeInstances function
func (m *MockEC2Client) DescribeInstances(input *ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error) {
	return &m.DescribeInstancesOutputValue, nil
}

func (m *MockEC2Client) StartInstances(input *ec2.StartInstancesInput) (*ec2.StartInstancesOutput, error) {
	return &m.StartInstancesOutputValue, nil
}

func (m mockStartInst) startInst(instId string, region string) (string, error) {
	return "i-0c938b5e573fb0f26", nil
}

func TestInstInfo(t *testing.T) {
	//Becauase we use the filter for included in the DescribeInstances we're testing to see
	//we get a valid data type back

	//Response from ec2 aws api and the expected result for function instInfo
	descInstancesWithStartWeekEnd := ec2.DescribeInstancesOutput{
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
								Key:   aws.String("StartWeekEnd"),
								Value: aws.String("15:00"),
							},
							{
								Key:   aws.String("StopWeekEnd"),
								Value: aws.String("17:00"),
							},
						},
					},
				},
			},
		},
	}

	cases := []struct {
		Name     string
		Resp     ec2.DescribeInstancesOutput
		Expected ec2.DescribeInstancesOutput
	}{
		{
			Name:     "StartWeekEndPresent",
			Resp:     descInstancesWithStartWeekEnd,
			Expected: descInstancesWithStartWeekEnd,
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
			inst, err := e.instInfo("StartWeekEnd", "us-east-1")

			if err != nil {
				fmt.Println("Unexpected Error - ", err.Error())
			}

			assert.Equal(t, "*ec2.DescribeInstancesOutput", fmt.Sprintf("%T", inst))

		})
	}
}

func TestStartInst(t *testing.T) {
	//Do we start the instance? What do we get back from the startInst function

	//The outputs we're mocking for the ec2 aws api for StartInstances
	var (
		//Output when starting Instance
		startInstancesOutput = ec2.StartInstancesOutput{
			StartingInstances: []*ec2.InstanceStateChange{
				{
					InstanceId: aws.String("i-0c61fb5e573fb0c55"),
					PreviousState: &ec2.InstanceState{
						Name: aws.String("stopped"),
					},
				},
			},
		}

		//Output when instance is stopped
		startInstancesOutputStopped = ec2.StartInstancesOutput{
			StartingInstances: []*ec2.InstanceStateChange{
				{
					InstanceId: aws.String("i-0c61fb5e573fb0c55"),
					PreviousState: &ec2.InstanceState{
						Name: aws.String("running"),
					},
				},
			},
		}
	)

	cases := []struct {
		Name     string
		Resp     ec2.StartInstancesOutput
		Expected string
	}{
		{
			Name:     "ProperInstanceId",
			Resp:     startInstancesOutput,
			Expected: "i-0c61fb5e573fb0c55",
		},
		{
			Name:     "InstanceRunning",
			Resp:     startInstancesOutputStopped,
			Expected: "i-0c61fb5e573fb0c55",
		},
	}

	for _, c := range cases {
		//Sub test for each case
		t.Run(c.Name, func(t *testing.T) {
			e := ec2Api{
				Client: &MockEC2Client{
					EC2API:                    nil,
					StartInstancesOutputValue: c.Resp,
				},
			}
			startInst, err := e.startInst("i-0c61fb5e573fb0c55", "us-east-1")

			if err != nil {
				t.Errorf("startInst(\"i-0c61fb5e573fb0c55\", \"us-east-1\") received %s; Expected %s", err.Error(), c.Expected)
			}

			assert.Equal(t, startInst, c.Expected)

		})
	}
}

func TestEvalInst(t *testing.T) {
	//Figure out what hapens with time, and day of the week, and if the instnace starts

	//Input to be used for eval inst. Perhaps pair it down to just the Instances section
	goodTime := ec2.DescribeInstancesOutput{
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
								Key:   aws.String("StartWeekEnd"),
								Value: aws.String("15:00"),
							},
							{
								Key:   aws.String("StopWeekEnd"),
								Value: aws.String("17:00"),
							},
						},
					},
				},
			},
		},
	}

	badTime := ec2.DescribeInstancesOutput{
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
								Key:   aws.String("StartWeekEnd"),
								Value: aws.String("155540:0"),
							},
							{
								Key:   aws.String("StopWeekEnd"),
								Value: aws.String("17:00"),
							},
						},
					},
				},
			},
		},
	}

	//badTag meaning autoscaling group to be found
	badTag := ec2.DescribeInstancesOutput{
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
								Key:   aws.String("StartWeekEnd"),
								Value: aws.String("15:00"),
							},
							{
								Key:   aws.String("StopWeekEnd"),
								Value: aws.String("17:00"),
							},
							{
								Key:   aws.String("aws:autoscaling:groupName"),
								Value: aws.String("testing-autoscaling-group"),
							},
						},
					},
				},
			},
		},
	}

	cases := []struct {
		Name     string
		Expected string
		Time     time.Time
		Input    *ec2.Instance
	}{
		{
			Name:     "StartInstance",
			Expected: "Starting Instance: i-0c938b5e573fb0f26",
			Time:     time.Date(2000, 11, 18, 15, 02, 00, 0, time.UTC),
			Input:    goodTime.Reservations[0].Instances[0],
		},
		{
			Name:     "Start time mismatch",
			Expected: "StartWeekEnd schedule not matched for: i-0c938b5e573fb0f26",
			Time:     time.Date(2000, 11, 18, 22, 02, 00, 0, time.UTC),
			Input:    goodTime.Reservations[0].Instances[0],
		},
		{
			Name:     "Weekday not Weekend",
			Expected: "Current day of week Friday. StartWeekEnd requires non weekday values",
			Time:     time.Date(2000, 11, 17, 22, 02, 00, 0, time.UTC),
			Input:    goodTime.Reservations[0].Instances[0],
		},
		{
			Name:     "Improper tag time format",
			Expected: "", //Error will throw instead
			Time:     time.Date(2000, 11, 18, 15, 02, 00, 0, time.UTC),
			Input:    badTime.Reservations[0].Instances[0],
		},
		{
			Name:     "Autoscaling group present",
			Expected: "Skipping - i-0c938b5e573fb0f26 is part of autoscaling group testing-autoscaling-group",
			Time:     time.Date(2000, 11, 18, 15, 02, 00, 0, time.UTC),
			Input:    badTag.Reservations[0].Instances[0],
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			//Put in the rest of the code to pass the mocked Start Inst
			e := mockStartInst{}
			result, err := evalInst(c.Input, "us-east-1", c.Time, e)
			if err != nil {
				assert.Equal(t, "parsing time \"155540:0\" as \"15:04\": cannot parse \"5540:0\" as \":\"", err.Error())
			}
			assert.Equal(t, c.Expected, result)
		})
	}
}
