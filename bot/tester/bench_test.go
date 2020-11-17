package tester

// todo finish
// todo clean up local bot.log afterwards
/*

func BenchmarkFullMessageHandling(b *testing.B) {
	cfg := config.Config{}
	logger := GetNullLogger()

	fakeSlack := StartFakeSlack(&cfg)

	bot := StartBot(cfg, logger)
	defer fakeSlack.Stop()

	kill := make(chan os.Signal, 1)
	go bot.HandleMessages(kill)

	for i := 0; i < b.N; i++ {
		fakeSlack.SendMessageToBot("#dev", "reply test")
	}

	kill<-syscall.Signal(1)
}

func BenchmarkStartBot(b *testing.B) {
	cfg := config.Config{}
	logger := GetNullLogger()

	fakeSlack := StartFakeSlack(&cfg)
	defer fakeSlack.Stop()

	var bot bot.Bot
	for i := 0; i < b.N; i++ {
		bot = StartBot(cfg, logger)
		time.Sleep(time.Millisecond * 10)
		bot.DisconnectRTM()
	}
}

func BenchmarkHandle(b *testing.B) {
	cfg := config.Config{}
	logger := GetNullLogger()

	fakeSlack := StartFakeSlack(&cfg)

	bot := StartBot(cfg, logger)
	//	defer bot.DisconnectRTM()
	defer fakeSlack.Stop()

	event := slack.MessageEvent{}
	event.User = "U123"
	event.Channel = "C1234"
	event.Text = "<@" + botID + "> reply test"

	for i := 0; i < b.N; i++ {
		bot.handleMessage(event)
	}
}
*/
