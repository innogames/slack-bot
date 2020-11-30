package games

import (
	"encoding/json"
	"fmt"
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/matcher"
	"github.com/innogames/slack-bot/bot/msg"
	"github.com/innogames/slack-bot/client"
	"github.com/pkg/errors"
	"html"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
)

const maxQuestions = 50 // api limit is 50
const apiURL = "https://opentdb.com/api.php"

func NewQuizCommand(slackClient client.SlackClient) bot.Command {
	return &quizCommand{slackClient: slackClient, apiURL: apiURL}
}

type quizCommand struct {
	slackClient client.SlackClient
	quiz        Quiz
	apiURL      string
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

func (c *quizCommand) GetMatcher() matcher.Matcher {
	return matcher.NewGroupMatcher(
		matcher.NewRegexpMatcher(`quiz`, c.StartQuiz),
		matcher.NewRegexpMatcher(`quiz (?P<questions>\d+)`, c.StartQuiz),
		matcher.NewRegexpMatcher(`answer (?P<answer>[\w\s]+)`, c.Answer),
	)
}

func (c *quizCommand) StartQuiz(match matcher.Result, message msg.Message) {
	questions := match.GetInt("questions")
	if questions == 0 {
		questions = 2
	}
	if questions > maxQuestions {
		c.slackClient.SendMessage(message, fmt.Sprintf("No more than %d questions allowed", maxQuestions))
		return
	}

	resp, err := http.Get(fmt.Sprintf("%s?amount=%d", c.apiURL, questions))
	if err != nil {
		c.slackClient.ReplyError(message, errors.Wrap(err, "Error while loading Quiz"))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			c.slackClient.ReplyError(message, errors.Wrap(err, "Error while loading Quiz"))
			return
		}

		if err := json.Unmarshal(bodyBytes, &c.quiz); err != nil {
			c.slackClient.ReplyError(message, errors.Wrap(err, "Error while loading Quiz"))
			return
		}
	}
	c.quiz.currentQuestion = 0
	c.quiz.tries = 0

	c.parseAnswers()

	c.printCurrentQuestion(message)
}

func (c *quizCommand) Answer(match matcher.Result, message msg.Message) {
	c.quiz.tries++

	answer := match.GetString("answer")
	if number, err := strconv.Atoi(answer); err == nil {
		answers := c.getCurrentAnswers()
		if number <= len(answers) {
			answer = answers[number-1]
		}
	}

	if c.getCurrentQuestion().CorrectAnswer == answer {
		c.slackClient.SendMessage(message, "correct")
		c.quiz.currentQuestion++
		if c.quiz.currentQuestion == len(c.quiz.Questions) {
			c.slackClient.SendMessage(message, fmt.Sprintf("You finished this quiz with %d Questions. You needed %d answers.", len(c.quiz.Questions), c.quiz.tries))
		} else {
			c.printCurrentQuestion(message)
		}
	} else {
		c.slackClient.SendMessage(message, "incorrect. try again")
	}
}

func (c *quizCommand) parseAnswers() {
	for questionNr, question := range c.quiz.Questions {
		answers := append(question.IncorrectAnswers, question.CorrectAnswer)

		rand.Shuffle(len(answers), func(i, j int) { answers[i], answers[j] = answers[j], answers[i] })

		c.quiz.Questions[questionNr].Answers = answers
	}
}

func (c *quizCommand) printCurrentQuestion(message msg.Message) {
	question := c.getCurrentQuestion()
	text := fmt.Sprintf(
		"Next question (#%d) is of *\"%s\" difficulty* from the category: \"*%s*\"\n",
		c.quiz.currentQuestion+1,
		html.UnescapeString(question.Difficulty),
		html.UnescapeString(question.Category),
	)
	text += html.UnescapeString(question.Question) + "\n"
	for index, answer := range question.Answers {
		text += fmt.Sprintf("%d.) %s\n", index+1, html.UnescapeString(answer))
	}

	text += ":interrobang: Hint type `answer {number}` to send your answer :interrobang:"

	c.slackClient.SendMessage(message, text)
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
			Category:    category,
			Examples: []string{
				"quiz",
				"quiz 10",
				"answer 2",
			},
		},
	}
}
