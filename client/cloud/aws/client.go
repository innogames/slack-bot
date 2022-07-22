package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

func GetSession() (*session.Session, error) {
	return session.NewSession(&aws.Config{
		Region: aws.String("ap-northeast-2"),
	})
}
