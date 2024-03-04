package sess

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
)

func EstablishSession(region string) (*session.Session, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			return nil, awsErr
		}
		return nil, err
	}

	return sess, nil
}
