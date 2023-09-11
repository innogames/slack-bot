package aws

import (
	"github.com/aws/aws-sdk-go/aws/session"
)

func getSession() (*session.Session, error) {
	return session.NewSession()
}
