package games

import (
	"encoding/json"
	"fmt"
	"github.com/nlopes/slack"
	"gitlab.innogames.de/foe-tools/slack-bot/bot"
	"gitlab.innogames.de/foe-tools/slack-bot/bot/matcher"
	"gitlab.innogames.de/foe-tools/slack-bot/client"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

func NewQuizCommand(slackClient client.SlackClient) bot.Command {
	return quizCommand{slackClient: slackClient}
}

type quizCommand struct {
	slackClient client.SlackClient
	quiz Quiz
}

type question struct {
	Category         string   `json:"category"`
	CorrectAnswer    string   `json:"correct_answer"`
	Difficulty       string   `json:"difficulty"`
	IncorrectAnswers []string `json:"incorrect_answers"`
	Question         string   `json:"question"`
	Type             string   `json:"type"`
	Answers          []string
}

type Quiz struct {
	ResponseCode int `json:"response_code"`
	Questions    []question `json:"results"`
	currentQuestion int
	tries        int
}


func (c quizCommand) GetMatcher() matcher.Matcher {
	return matcher.NewGroupMatcher(
		matcher.NewRegexpMatcher(`quiz`, c.StartQuiz),
		matcher.NewRegexpMatcher(`answer (?P<answer>[\w\s]+)`, c.Answer),
	)
}

func (c *quizCommand) StartQuiz(match matcher.Result, event slack.MessageEvent) {

	resp, _ := http.Get("https://opentdb.com/api.php?amount=2")

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		if err := json.Unmarshal(bodyBytes, &c.quiz); err != nil {
			panic(err)
		}
	}

	c.parseAnswers()

	c.printCurrentQuestion(event)
}

func (c *quizCommand) Answer(match matcher.Result, event slack.MessageEvent) {
	c.quiz.tries++

	answer := match.GetString("answer")
	if number, err := strconv.Atoi(answer); err == nil {
		answers := c.getCurrentAnswers()
		if number <= len(answers) {
			answer = answers[number-1]
		}
	}

	if c.getCurrentQuestion().CorrectAnswer == answer {
		c.slackClient.Reply(event, "correct")
		c.quiz.currentQuestion++
		if c.quiz.currentQuestion == len(c.quiz.Questions) {
			c.slackClient.Reply(event, fmt.Sprintf("You finished our quiz with %d Questions. You needed %d tries.", len(c.quiz.Questions), c.quiz.tries))
		} else {
			c.printCurrentQuestion(event)
		}
	} else {
		c.slackClient.Reply(event, "incorrect. try again")
	}
}

func (c *quizCommand) parseAnswers() {
	for questionNr, question := range c.quiz.Questions {
		answers := append(question.IncorrectAnswers, question.CorrectAnswer)

		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(len(answers), func(i, j int) { answers[i], answers[j] = answers[j], answers[i] })

		c.quiz.Questions[questionNr].Answers = answers
	}
}

func (c *quizCommand) printCurrentQuestion(event slack.MessageEvent) {
	question := c.getCurrentQuestion()
	c.slackClient.Reply(event, fmt.Sprintf("Next question is %s from the category: %s", question.Difficulty, question.Category))
	c.slackClient.Reply(event, question.Question)
	for index, answer := range question.Answers {
		c.slackClient.Reply(event, strconv.Itoa(index + 1) + ".) " + answer)
	}
}

func (c *quizCommand) getCurrentQuestion() question {
	return c.quiz.Questions[c.quiz.currentQuestion]
}

func (c *quizCommand) getCurrentAnswers() []string {
	return append(c.getCurrentQuestion().Answers)
}
