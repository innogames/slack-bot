package aws

import (
	"context"
	"fmt"
	"log"
	"os"
	"path"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/matcher"
	"github.com/innogames/slack-bot/v2/bot/msg"
)

// command to trigger/start jenkins jobs
type ecsCommand struct {
	awsCommand
}

var _ecs *ecs.Client

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
	ctx := context.Background()
	svc := assertECS()

	params := &ecs.ListServicesInput{Cluster: &cluster}
	paginator := ecs.NewListServicesPaginator(svc, params)

	result := make([]string, 0)
	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		for _, serviceArn := range output.ServiceArns {
			result = append(result, path.Base(serviceArn))
		}
	}

	log.Println(result)
	return result, nil
}

func assertECS() *ecs.Client {
	if _ecs == nil {
		ctx := context.Background()

		// Load AWS config with optional region override
		var cfg aws.Config
		var err error

		if region := os.Getenv("AWS_REGION"); region != "" {
			cfg, err = awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region))
		} else {
			cfg, err = awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion("us-east-1"))
		}

		if err != nil {
			log.Printf("Error loading AWS config: %v", err)
			return nil
		}

		log.Println("Using AWS region ", cfg.Region)
		_ecs = ecs.NewFromConfig(cfg)
	}
	return _ecs
}

func ForceNewDeployment(clusterName string, serviceName string) error {
	ctx := context.Background()

	// Load AWS config with optional region override
	var cfg aws.Config
	var err error

	if region := os.Getenv("AWS_REGION"); region != "" {
		cfg, err = awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region))
	} else {
		cfg, err = awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion("us-east-1"))
	}

	if err != nil {
		return err
	}

	ecsSvc := ecs.NewFromConfig(cfg)

	serviceParams := &ecs.DescribeServicesInput{
		Services: []string{serviceName},
		Cluster:  &clusterName,
	}
	result, err := ecsSvc.DescribeServices(ctx, serviceParams)
	if err != nil {
		return err
	}
	if len(result.Services) == 0 {
		return fmt.Errorf("could not find service %s in cluster %s", serviceName, clusterName)
	}

	// Update Service
	for i := range result.Services {
		service := result.Services[i]
		if *service.ServiceName == serviceName {
			newServiceParams := &ecs.UpdateServiceInput{
				Service:            service.ServiceName,
				Cluster:            &clusterName,
				ForceNewDeployment: true,
			}
			_, err := ecsSvc.UpdateService(ctx, newServiceParams)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
