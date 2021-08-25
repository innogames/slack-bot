package jira

import (
	"strings"
	"testing"

	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/bot/util"
	"github.com/innogames/slack-bot/v2/client"
	"github.com/innogames/slack-bot/v2/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestJiraSearch(t *testing.T) {
	slackClient := &mocks.SlackClient{}

	cfg := &config.Jira{
		Host:    "https://issues.apache.org/jira/",
		Project: "ZOOKEEPER",
	}
	jiraClient, err := client.GetJiraClient(cfg)
	assert.Nil(t, err)

	command := bot.Commands{}
	command.AddCommand(newJiraCommand(jiraClient, slackClient, cfg))

	t.Run("No match", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "quatsch"

		actual := command.Run(message)
		assert.False(t, actual)
	})

	t.Run("search existing ticket", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "jira ZOOKEEPER-3456"

		expected := "[{\"color\":\"#D3D3D3\",\"fields\":[{\"title\":\"Name\",\"value\":\"\\u003chttps://issues.apache.org/jira/browse/ZOOKEEPER-3456|ZOOKEEPER-3456\\u003e: Service temporarily unavailable due to an ongoing leader election. Please refresh\",\"short\":false},{\"title\":\"Priority\",\"value\":\"Major\",\"short\":true},{\"title\":\"Type\",\"value\":\"Bug\",\"short\":true},{\"title\":\"Status\",\"value\":\"Open\",\"short\":true},{\"title\":\"Components\",\"value\":\"server\",\"short\":true},{\"title\":\"Description\",\"value\":\"Hi\\r\\n\\r\\nI configured Zookeeper with four nodes for my Mesos cluster with Marathon. When I ran Flink Json file on Marathon, it was run without problem. But, when I entered IP of my two slaves, just one slave shew Flink UI and another slave shew this error:\\r\\n\\r\\n\u00a0\\r\\n\\r\\nService temporarily unavailable due to an ongoing leader election. Please refresh\\r\\n\\r\\nI checked \\\"zookeeper.out\\\" file and it said that :\\r\\n\\r\\n\u00a0\\r\\n\\r\\n019-07-07 11:48:43,412 [myid:] - INFO [main:QuorumPeerConfig@136] - Reading configuration from: /home/zookeeper-3.4.14/bin/../conf/zoo.cfg\\r\\n2019-07-07 11:48:43,421 [myid:] - INFO [main:QuorumPeer$QuorumServer@185] - Resolved hostname: 0.0.0.0 to address: /0.0.0.0\\r\\n2019-07-07 11:48:43,421 [myid:] - INFO [main:QuorumPeer$QuorumServer@185] - Resolved hostname: 10.32.0.3 to address: /10.32.0.3\\r\\n2019-07-07 11:48:43,422 [myid:] - INFO [main:QuorumPeer$QuorumServer@185] - Resolved hostname: 10.32.0.2 to address: /10.32.0.2\\r\\n2019-07-07 11:48:43,422 [myid:] - INFO [main:QuorumPeer$QuorumServer@185] - Resolved hostname: 10.32.0.5 to address: /10.32.0.5\\r\\n2019-07-07 11:48:43,422 [myid:] - WARN [main:QuorumPeerConfig@354] - Non-optimial configuration, consider an odd number of servers.\\r\\n2019-07-07 11:48:43,422 [myid:] - INFO [main:QuorumPeerConfig@398] - Defaulting to majority quorums\\r\\n2019-07-07 11:48:43,425 [myid:3] - INFO [main:DatadirCleanupManager@78] - autopurge.snapRetainCount set to 3\\r\\n2019-07-07 11:48:43,425 [myid:3] - INFO [main:DatadirCleanupManager@79] - autopurge.purgeInterval set to 0\\r\\n2019-07-07 11:48:43,425 [myid:3] - INFO [main:DatadirCleanupManager@101] - Purge task is not scheduled.\\r\\n2019-07-07 11:48:43,432 [myid:3] - INFO [main:QuorumPeerMain@130] - Starting quorum peer\\r\\n2019-07-07 11:48:43,437 [myid:3] - INFO [main:ServerCnxnFactory@117] - Using org.apache.zookeeper.server.NIOServerCnxnFactory as server connect$\\r\\n2019-07-07 11:48:43,439 [myid:3] - INFO [main:NIOServerCnxnFactory@89] - binding to port 0.0.0.0/0.0.0.0:2181\\r\\n2019-07-07 11:48:43,440 [myid:3] - ERROR [main:QuorumPeerMain@92] - Unexpected exception, exiting abnormally\\r\\njava.net.BindException: Address already in use\\r\\n at sun.nio.ch.Net.bind0(Native Method)\\r\\n at sun.nio.ch.Net.bind(Net.java:433)\\r\\n at sun.nio.ch.Net.bind(Net.java:425)\\r\\n at sun.nio.ch.ServerSocketChannelImpl.bind(ServerSocketChannelImpl.java:223)\\r\\n at sun.nio.ch.ServerSocketAdaptor.bind(ServerSocketAdaptor.java:74)\\r\\n at sun.nio.ch.ServerSocketAdaptor.bind(ServerSocketAdaptor.java:67)\\r\\n at org.apache.zookeeper.server.NIOServerCnxnFactory.configure(NIOServerCnxnFactory.java:90)\\r\\n at org.apache.zookeeper.server.quorum.QuorumPeerMain.runFromConfig(QuorumPeerMain.java:133)\\r\\n at org.apache.zookeeper.server.quorum.QuorumPeerMain.initializeAndRun(QuorumPeerMain.java:114)\\r\\n at org.apache.zookeeper.server.quorum.QuorumPeerMain.main(QuorumPeerMain.java:81)\\r\\n\\r\\n\u00a0\\r\\n\\r\\nI searched a lot and could not find the solution.\",\"short\":false}],\"actions\":[{\"name\":\"\",\"text\":\"Open in Jira\",\"style\":\"default\",\"type\":\"button\",\"url\":\"https://issues.apache.org/jira/browse/ZOOKEEPER-3456\"}],\"mrkdwn_in\":[\"text\",\"fields\"],\"blocks\":null}]"
		mocks.AssertSlackJSON(t, slackClient, message, expected)

		actual := command.Run(message)
		assert.True(t, actual)
	})

	t.Run("search existing ticket of default project", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "jira 3456"

		expected := "[{\"color\":\"#D3D3D3\",\"fields\":[{\"title\":\"Name\",\"value\":\"\\u003chttps://issues.apache.org/jira/browse/ZOOKEEPER-3456|ZOOKEEPER-3456\\u003e: Service temporarily unavailable due to an ongoing leader election. Please refresh\",\"short\":false},{\"title\":\"Priority\",\"value\":\"Major\",\"short\":true},{\"title\":\"Type\",\"value\":\"Bug\",\"short\":true},{\"title\":\"Status\",\"value\":\"Open\",\"short\":true},{\"title\":\"Components\",\"value\":\"server\",\"short\":true},{\"title\":\"Description\",\"value\":\"Hi\\r\\n\\r\\nI configured Zookeeper with four nodes for my Mesos cluster with Marathon. When I ran Flink Json file on Marathon, it was run without problem. But, when I entered IP of my two slaves, just one slave shew Flink UI and another slave shew this error:\\r\\n\\r\\n\u00a0\\r\\n\\r\\nService temporarily unavailable due to an ongoing leader election. Please refresh\\r\\n\\r\\nI checked \\\"zookeeper.out\\\" file and it said that :\\r\\n\\r\\n\u00a0\\r\\n\\r\\n019-07-07 11:48:43,412 [myid:] - INFO [main:QuorumPeerConfig@136] - Reading configuration from: /home/zookeeper-3.4.14/bin/../conf/zoo.cfg\\r\\n2019-07-07 11:48:43,421 [myid:] - INFO [main:QuorumPeer$QuorumServer@185] - Resolved hostname: 0.0.0.0 to address: /0.0.0.0\\r\\n2019-07-07 11:48:43,421 [myid:] - INFO [main:QuorumPeer$QuorumServer@185] - Resolved hostname: 10.32.0.3 to address: /10.32.0.3\\r\\n2019-07-07 11:48:43,422 [myid:] - INFO [main:QuorumPeer$QuorumServer@185] - Resolved hostname: 10.32.0.2 to address: /10.32.0.2\\r\\n2019-07-07 11:48:43,422 [myid:] - INFO [main:QuorumPeer$QuorumServer@185] - Resolved hostname: 10.32.0.5 to address: /10.32.0.5\\r\\n2019-07-07 11:48:43,422 [myid:] - WARN [main:QuorumPeerConfig@354] - Non-optimial configuration, consider an odd number of servers.\\r\\n2019-07-07 11:48:43,422 [myid:] - INFO [main:QuorumPeerConfig@398] - Defaulting to majority quorums\\r\\n2019-07-07 11:48:43,425 [myid:3] - INFO [main:DatadirCleanupManager@78] - autopurge.snapRetainCount set to 3\\r\\n2019-07-07 11:48:43,425 [myid:3] - INFO [main:DatadirCleanupManager@79] - autopurge.purgeInterval set to 0\\r\\n2019-07-07 11:48:43,425 [myid:3] - INFO [main:DatadirCleanupManager@101] - Purge task is not scheduled.\\r\\n2019-07-07 11:48:43,432 [myid:3] - INFO [main:QuorumPeerMain@130] - Starting quorum peer\\r\\n2019-07-07 11:48:43,437 [myid:3] - INFO [main:ServerCnxnFactory@117] - Using org.apache.zookeeper.server.NIOServerCnxnFactory as server connect$\\r\\n2019-07-07 11:48:43,439 [myid:3] - INFO [main:NIOServerCnxnFactory@89] - binding to port 0.0.0.0/0.0.0.0:2181\\r\\n2019-07-07 11:48:43,440 [myid:3] - ERROR [main:QuorumPeerMain@92] - Unexpected exception, exiting abnormally\\r\\njava.net.BindException: Address already in use\\r\\n at sun.nio.ch.Net.bind0(Native Method)\\r\\n at sun.nio.ch.Net.bind(Net.java:433)\\r\\n at sun.nio.ch.Net.bind(Net.java:425)\\r\\n at sun.nio.ch.ServerSocketChannelImpl.bind(ServerSocketChannelImpl.java:223)\\r\\n at sun.nio.ch.ServerSocketAdaptor.bind(ServerSocketAdaptor.java:74)\\r\\n at sun.nio.ch.ServerSocketAdaptor.bind(ServerSocketAdaptor.java:67)\\r\\n at org.apache.zookeeper.server.NIOServerCnxnFactory.configure(NIOServerCnxnFactory.java:90)\\r\\n at org.apache.zookeeper.server.quorum.QuorumPeerMain.runFromConfig(QuorumPeerMain.java:133)\\r\\n at org.apache.zookeeper.server.quorum.QuorumPeerMain.initializeAndRun(QuorumPeerMain.java:114)\\r\\n at org.apache.zookeeper.server.quorum.QuorumPeerMain.main(QuorumPeerMain.java:81)\\r\\n\\r\\n\u00a0\\r\\n\\r\\nI searched a lot and could not find the solution.\",\"short\":false}],\"actions\":[{\"name\":\"\",\"text\":\"Open in Jira\",\"style\":\"default\",\"type\":\"button\",\"url\":\"https://issues.apache.org/jira/browse/ZOOKEEPER-3456\"}],\"mrkdwn_in\":[\"text\",\"fields\"],\"blocks\":null}]"
		mocks.AssertSlackJSON(t, slackClient, message, expected)

		actual := command.Run(message)
		assert.True(t, actual)
	})

	t.Run("search existing ticket by slack link", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "https://issues.apache.org/jira/browse/ZOOKEEPER-3456"

		expected := "[{\"color\":\"#D3D3D3\",\"fields\":[{\"title\":\"Name\",\"value\":\"\\u003chttps://issues.apache.org/jira/browse/ZOOKEEPER-3456|ZOOKEEPER-3456\\u003e: Service temporarily unavailable due to an ongoing leader election. Please refresh\",\"short\":false},{\"title\":\"Priority\",\"value\":\"Major\",\"short\":true},{\"title\":\"Type\",\"value\":\"Bug\",\"short\":true},{\"title\":\"Status\",\"value\":\"Open\",\"short\":true},{\"title\":\"Components\",\"value\":\"server\",\"short\":true},{\"title\":\"Description\",\"value\":\"Hi\\r\\n\\r\\nI configured Zookeeper with four nodes for my Mesos cluster with Marathon. When I ran Flink Json file on Marathon, it was run without problem. But, when I entered IP of my two slaves, just one slave shew Flink UI and another slave shew this error:\\r\\n\\r\\n\u00a0\\r\\n\\r\\nService temporarily unavailable due to an ongoing leader election. Please refresh\\r\\n\\r\\nI checked \\\"zookeeper.out\\\" file and it said that :\\r\\n\\r\\n\u00a0\\r\\n\\r\\n019-07-07 11:48:43,412 [myid:] - INFO [main:QuorumPeerConfig@136] - Reading configuration from: /home/zookeeper-3.4.14/bin/../conf/zoo.cfg\\r\\n2019-07-07 11:48:43,421 [myid:] - INFO [main:QuorumPeer$QuorumServer@185] - Resolved hostname: 0.0.0.0 to address: /0.0.0.0\\r\\n2019-07-07 11:48:43,421 [myid:] - INFO [main:QuorumPeer$QuorumServer@185] - Resolved hostname: 10.32.0.3 to address: /10.32.0.3\\r\\n2019-07-07 11:48:43,422 [myid:] - INFO [main:QuorumPeer$QuorumServer@185] - Resolved hostname: 10.32.0.2 to address: /10.32.0.2\\r\\n2019-07-07 11:48:43,422 [myid:] - INFO [main:QuorumPeer$QuorumServer@185] - Resolved hostname: 10.32.0.5 to address: /10.32.0.5\\r\\n2019-07-07 11:48:43,422 [myid:] - WARN [main:QuorumPeerConfig@354] - Non-optimial configuration, consider an odd number of servers.\\r\\n2019-07-07 11:48:43,422 [myid:] - INFO [main:QuorumPeerConfig@398] - Defaulting to majority quorums\\r\\n2019-07-07 11:48:43,425 [myid:3] - INFO [main:DatadirCleanupManager@78] - autopurge.snapRetainCount set to 3\\r\\n2019-07-07 11:48:43,425 [myid:3] - INFO [main:DatadirCleanupManager@79] - autopurge.purgeInterval set to 0\\r\\n2019-07-07 11:48:43,425 [myid:3] - INFO [main:DatadirCleanupManager@101] - Purge task is not scheduled.\\r\\n2019-07-07 11:48:43,432 [myid:3] - INFO [main:QuorumPeerMain@130] - Starting quorum peer\\r\\n2019-07-07 11:48:43,437 [myid:3] - INFO [main:ServerCnxnFactory@117] - Using org.apache.zookeeper.server.NIOServerCnxnFactory as server connect$\\r\\n2019-07-07 11:48:43,439 [myid:3] - INFO [main:NIOServerCnxnFactory@89] - binding to port 0.0.0.0/0.0.0.0:2181\\r\\n2019-07-07 11:48:43,440 [myid:3] - ERROR [main:QuorumPeerMain@92] - Unexpected exception, exiting abnormally\\r\\njava.net.BindException: Address already in use\\r\\n at sun.nio.ch.Net.bind0(Native Method)\\r\\n at sun.nio.ch.Net.bind(Net.java:433)\\r\\n at sun.nio.ch.Net.bind(Net.java:425)\\r\\n at sun.nio.ch.ServerSocketChannelImpl.bind(ServerSocketChannelImpl.java:223)\\r\\n at sun.nio.ch.ServerSocketAdaptor.bind(ServerSocketAdaptor.java:74)\\r\\n at sun.nio.ch.ServerSocketAdaptor.bind(ServerSocketAdaptor.java:67)\\r\\n at org.apache.zookeeper.server.NIOServerCnxnFactory.configure(NIOServerCnxnFactory.java:90)\\r\\n at org.apache.zookeeper.server.quorum.QuorumPeerMain.runFromConfig(QuorumPeerMain.java:133)\\r\\n at org.apache.zookeeper.server.quorum.QuorumPeerMain.initializeAndRun(QuorumPeerMain.java:114)\\r\\n at org.apache.zookeeper.server.quorum.QuorumPeerMain.main(QuorumPeerMain.java:81)\\r\\n\\r\\n\u00a0\\r\\n\\r\\nI searched a lot and could not find the solution.\",\"short\":false}],\"actions\":[{\"name\":\"\",\"text\":\"Open in Jira\",\"style\":\"default\",\"type\":\"button\",\"url\":\"https://issues.apache.org/jira/browse/ZOOKEEPER-3456\"}],\"mrkdwn_in\":[\"text\",\"fields\"],\"blocks\":null}]"
		mocks.AssertSlackJSON(t, slackClient, message, expected)

		actual := command.Run(message)
		assert.True(t, actual)
	})

	t.Run("print ticket link", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "jira link ZOOKEEPER-3455"

		slackClient.On("SendMessage", message, "<https://issues.apache.org/jira/browse/ZOOKEEPER-3455|ZOOKEEPER-3455: Java 13 build failure on trunk: UnifiedServerSocketTest.testConnectWithoutSSLToStrictServer>").Return("")

		actual := command.Run(message)
		assert.True(t, actual)
	})

	t.Run("search invalid ticket", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "jira ZOOKEEPER-10000000000"

		slackClient.On("SendMessage", message, "Issue Does Not Exist: request failed. Please analyze the request body for more details. Status code: 404").Return("")

		actual := command.Run(message)
		assert.True(t, actual)
	})

	t.Run("search invalid JQL", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "jql FOO=BAR"

		slackClient.On(
			"SendMessage",
			message,
			mock.MatchedBy(func(message string) bool {
				return strings.HasPrefix(message, "Field 'FOO' does not exist or")
			})).Return("")
		actual := command.Run(message)
		assert.True(t, actual)
	})

	t.Run("Render template with jiraTicket()", func(t *testing.T) {
		tpl, err := util.CompileTemplate(`{{$ticket := jiraTicket "ZOOKEEPER-3455"}}ID: {{$ticket.ID}} Key: {{$ticket.Key}}`)
		assert.Nil(t, err)

		res, err := util.EvalTemplate(tpl, util.Parameters{})
		assert.Nil(t, err)

		assert.Equal(t, "ID: 13242741 Key: ZOOKEEPER-3455", res)
	})

	t.Run("Render template with jiraTicketUrl()", func(t *testing.T) {
		tpl, err := util.CompileTemplate(`{{ jiraTicketUrl "ZOOKEEPER-3455"}}`)
		assert.Nil(t, err)

		res, err := util.EvalTemplate(tpl, util.Parameters{})
		assert.Nil(t, err)

		assert.Equal(t, "https://issues.apache.org/jira/browse/ZOOKEEPER-3455", res)
	})

	t.Run("Test help", func(t *testing.T) {
		help := command.GetHelp()
		assert.GreaterOrEqual(t, len(help), 1)
	})
}

func TestConvertMarkdown(t *testing.T) {
	message := "h1. hallo how are {code}you{code}?"
	actual := convertMarkdown(message)

	assert.Equal(t, "hallo how are ```you```?", actual)
}

func BenchmarkConvertMarkdown(b *testing.B) {
	message1 := "h1. hallo how are {code}you{code}?"
	message2 := "hallo how are you?"

	for i := 0; i < b.N; i++ {
		convertMarkdown(message1)
		convertMarkdown(message2)
	}
}
