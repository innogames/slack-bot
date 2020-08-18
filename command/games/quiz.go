package games

import (
	"encoding/json"
	"fmt"
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/matcher"
	"github.com/innogames/slack-bot/client"
	"github.com/slack-go/slack"
	"github.com/pkg/errors"
	"html"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
)

const maxQuestions int = 50 // api limit is 50
const apiUrl string = "https://opentdb.com/api.php"

func NewQuizCommand(slackClient client.SlackClient) *quizCommand {
	return &quizCommand{slackClient: slackClient, apiUrl: apiUrl}
}

type quizCommand struct {
	slackClient client.SlackClient
	quiz        Quiz
	apiUrl      string
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
	ResponseCode    int        `json:"response_code"`
	Questions       []question `json:"results"`
	currentQuestion int
	tries           int
}

func (c quizCommand) GetMatcher() matcher.Matcher {
	return matcher.NewGroupMatcher(
		matcher.NewRegexpMatcher(`quiz`, c.StartQuiz),
		matcher.NewRegexpMatcher(`quiz (?P<questions>\d+)`, c.StartQuiz),
		matcher.NewRegexpMatcher(`answer (?P<answer>[\w\s]+)`, c.Answer),
	)
}

func (c *quizCommand) StartQuiz(match matcher.Result, event slack.MessageEvent) {
	questions := match.GetInt("questions")
	if questions == 0 {
		questions = 2
	}
	if questions > maxQuestions {
		c.slackClient.Reply(event, fmt.Sprintf("No more than %d questions allowed", maxQuestions))
		return
	}

	resp, err := http.Get(fmt.Sprintf("%s?amount=%d", c.apiUrl, questions))
	if err != nil {
		c.slackClient.ReplyError(event, errors.Wrap(err, "Error while loading Quiz"))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			c.slackClient.ReplyError(event, errors.Wrap(err, "Error while loading Quiz"))
			return
		}

		if err := json.Unmarshal(bodyBytes, &c.quiz); err != nil {
			c.slackClient.ReplyError(event, errors.Wrap(err, "Error while loading Quiz"))
			return
		}
	}
	c.quiz.currentQuestion = 0
	c.quiz.tries = 0

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
			c.slackClient.Reply(event, fmt.Sprintf("You finished this quiz with %d Questions. You needed %d answers.", len(c.quiz.Questions), c.quiz.tries))
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

		rand.Shuffle(len(answers), func(i, j int) { answers[i], answers[j] = answers[j], answers[i] })

		c.quiz.Questions[questionNr].Answers = answers
	}
}

func (c *quizCommand) printCurrentQuestion(event slack.MessageEvent) {
	question := c.getCurrentQuestion()
	message := fmt.Sprintf(
		"Next question (#%d) is of *\"%s\" difficulty* from the category: \"*%s*\"\n",
		c.quiz.currentQuestion+1,
		html.UnescapeString(question.Difficulty),
		html.UnescapeString(question.Category),
	)
	message += html.UnescapeString(question.Question) + "\n"
	for index, answer := range question.Answers {
		message += fmt.Sprintf("%d.) %s\n", index+1, html.UnescapeString(answer))
	}
	message += ":interrobang: Hint type `answer {number}` to send your answer :interrobang:"

	c.slackClient.Reply(event, message)
}

func (c *quizCommand) getCurrentQuestion() question {
	return c.quiz.Questions[c.quiz.currentQuestion]
}

func (c *quizCommand) getCurrentAnswers() []string {
	return c.getCurrentQuestion().Answers
}

func (c *quizCommand) GetHelp() []bot.Help {
	return []bot.Help{
		{
			Command:     "quiz",
			Description: "small quiz for a nice break",
			Examples: []string{
				"quiz",
				"quiz 10",
				"answer 2",
			},
		},
	}
}
