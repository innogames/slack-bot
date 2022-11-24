package aws

import (
	"fmt"
	"log"
	"os"
	"path"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/matcher"
	"github.com/innogames/slack-bot/v2/bot/msg"
)

// command to trigger/start jenkins jobs
type ecsCommand struct {
	awsCommand
}

var _ecs *ecs.ECS

// NewAwsCommand is a command to interact with aws resources
func newEcsCommands(base awsCommand) bot.Command {
	return &ecsCommand{base}
}

func (c *ecsCommand) GetMatcher() matcher.Matcher {
	return matcher.NewGroupMatcher(
		matcher.NewRegexpMatcher(`ecs ls (?P<CLUSTER>.*)`, c.list),
		matcher.NewRegexpMatcher(`ecs restart (?P<CLUSTER>.*) (?P<SERVICE>.*)`, c.restart),
	)
}

// We want to retrieve all services from an ecs cluster and throw them as a slack block

func (c *ecsCommand) list(match matcher.Result, message msg.Message) {
	cluster := match.GetString("CLUSTER")
	svc, err := ListServices(cluster)
	if err != nil {
		log.Println(err.Error())
		c.SendEphemeralMessage(message, err.Error())
		return
	}

	var text string

	text += "Hello <@" + message.User + ">, current services:. \n"

	for _, v := range svc {
		text += fmt.Sprintf("• %s", v)
		text += "\n"
	}

	c.SendEphemeralMessage(message, text)
}

// restart service

func (c *ecsCommand) restart(match matcher.Result, message msg.Message) {
	cluster := match.GetString("CLUSTER")
	svc := match.GetString("SERVICE")
	log.Println("restart requested: ", svc)
	err := ForceNewDeployment(cluster, svc)

	if err != nil {
		log.Println(err.Error())
		c.SendEphemeralMessage(message, err.Error())
		return
	}

	// TODO let user know when restart is completed
	// message := fmt.Sprintf("restart completed")
	c.SendMessage(message, svc)
}

func (c *ecsCommand) GetHelp() []bot.Help {
	examples := []string{
		"ecs ls my-cluster-name // to list services in a cluster",
		"ecs restart my-cluster-name service // to restart a service in a cluster",
	}

	help := make([]bot.Help, 0)
	help = append(help, bot.Help{
		Command:     "ecs <sub command>",
		Description: "interact with aws ecs resources",
		Examples:    examples,
		Category:    category,
	})

	return help
}

func ListServices(cluster string) ([]string, error) {
	svc := assertECS()
	params := &ecs.ListServicesInput{Cluster: aws.String(cluster)}
	result := make([]*string, 0)
	err := svc.ListServicesPages(params, func(services *ecs.ListServicesOutput, lastPage bool) bool {
		result = append(result, services.ServiceArns...)
		return !lastPage
	})
	if err != nil {
		return nil, err
	}
	out := make([]string, len(result))
	for i, s := range result {
		out[i] = path.Base(*s)
	}
	log.Println(out)
	return out, nil
}

func assertECS() *ecs.ECS {
	if _ecs == nil {
		_ecs = ecs.New(session.New(getServiceConfiguration()))
	}
	return _ecs
}

func getServiceConfiguration() *aws.Config {
	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = "us-east-1"
	}
	log.Println("Using AWS region ", region)
	return &aws.Config{Region: aws.String(region)}
}

func ForceNewDeployment(clusterName string, serviceName string) error {
	sess, err := session.NewSession()
	if err != nil {
		return err
	}
	ecsSvc := ecs.New(sess, &aws.Config{Region: aws.String(os.Getenv("AWS_REGION"))})

	serviceParams := &ecs.DescribeServicesInput{
		Services: []*string{
			aws.String(serviceName),
		},
		Cluster: aws.String(clusterName),
	}
	result, err := ecsSvc.DescribeServices(serviceParams)
	if err != nil {
		return err
	}
	if len(result.Services) == 0 {
		return fmt.Errorf("Could not find service %s in cluster %s", serviceName, clusterName)
	}

	// Update Service
	for i := range result.Services {
		service := result.Services[i]
		if *service.ServiceName == serviceName {
			newServiceParams := &ecs.UpdateServiceInput{
				Service:            service.ServiceName,
				Cluster:            aws.String(clusterName),
				ForceNewDeployment: aws.Bool(true),
			}
			_, err := ecsSvc.UpdateService(newServiceParams)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
