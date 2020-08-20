// Package formx provides user-editable forms
package formx

import (
	"crypto/sha1"
	"database/sql/driver"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"net/url"
	"strings"

	"github.com/microcosm-cc/bluemonday"
)

type Option struct {
	Value   string `json:"Value"`
	Display string `json:"Display"`
}

type Subquestion struct {
	Name string `json:"Name"`
	Text string `json:"Text"`
}

type Question struct {
	Type         string        `json:"Type"`
	Text         string        `json:"Text"`
	Name         string        `json:"Name"`
	Options      []Option      `json:"Options"`
	Subquestions []Subquestion `json:"Subquestions"`
}

func (question Question) Value() (driver.Value, error) {
	b, err := json.Marshal(question)
	return driver.Value(string(b)), err
}

func (question *Question) Scan(value interface{}) (err error) {
	switch v := value.(type) {
	case nil:
		return nil
	case string:
		err = json.Unmarshal([]byte(v), &question)
		return err
	case []byte:
		err = json.Unmarshal(v, &question)
		return err
	default:
		return fmt.Errorf("value %#v from database is neither a string nor NULL", value)
	}
}

type Questions []Question

func (questions Questions) Value() (driver.Value, error) {
	b, err := json.Marshal(questions)
	return driver.Value(string(b)), err
}

func (questions *Questions) Scan(value interface{}) error {
	var err error
	switch v := value.(type) {
	case nil:
		return nil
	case string:
		err = json.Unmarshal([]byte(v), questions)
	case []byte:
		err = json.Unmarshal(v, questions)
	default:
		return fmt.Errorf("value %#v from database is neither a string nor NULL", value)
	}
	return err
}

type Answers map[string][]string

func (answers Answers) Value() (driver.Value, error) {
	b, err := json.Marshal(answers)
	return driver.Value(string(b)), err
}

func (answers *Answers) Scan(value interface{}) error {
	switch v := value.(type) {
	case nil:
		return nil
	case string:
		return json.Unmarshal([]byte(v), answers)
	case []byte:
		return json.Unmarshal(v, answers)
	default:
		return fmt.Errorf("value %#v from database is neither a string nor NULL", value)
	}
}

func (answers Answers) IsEmpty() bool {
	for i := range answers {
		if len(answers[i]) != 0 {
			return false
		}
	}
	return true
}

type SubquestionAnswer struct {
	Name   string `json:"Name"`
	Text   string `json:"Text"`
	Answer string `json:"Answer"`
}

type QuestionAnswer struct {
	Type               string              `json:"Type"`
	Text               string              `json:"Text"`
	Name               string              `json:"Name"`
	Options            []Option            `json:"Options"`
	SubquestionAnswers []SubquestionAnswer `json:"SubquestionAnswers"`
	Answer             []string            `json:"Answer"`
}

const (
	QuestionTypeParagraph  = "paragraph"
	QuestionTypeShorttext  = "short text"
	QuestionTypeLongtext   = "long text"
	QuestionTypeCheckbox   = "checkbox"
	QuestionTypeSelect     = "select"
	QuestionTypeRadio      = "radio"
	QuestionTypeMultiradio = "multiradio"
	QuestionTypeDate       = "date"
	QuestionTypeTime       = "time"
	QuestionTypeImage      = "image"
	QuestionTypeNull       = ""
)

func addQuestionTypes(funcs template.FuncMap) template.FuncMap {
	if funcs == nil {
		funcs = template.FuncMap{}
	}
	funcs["QuestionTypeParagraph"] = func() string { return QuestionTypeParagraph }
	funcs["QuestionTypeShorttext"] = func() string { return QuestionTypeShorttext }
	funcs["QuestionTypeLongtext"] = func() string { return QuestionTypeLongtext }
	funcs["QuestionTypeCheckbox"] = func() string { return QuestionTypeCheckbox }
	funcs["QuestionTypeSelect"] = func() string { return QuestionTypeSelect }
	funcs["QuestionTypeRadio"] = func() string { return QuestionTypeRadio }
	funcs["QuestionTypeMultiradio"] = func() string { return QuestionTypeMultiradio }
	funcs["QuestionTypeDate"] = func() string { return QuestionTypeDate }
	funcs["QuestionTypeTime"] = func() string { return QuestionTypeTime }
	funcs["QuestionTypeImage"] = func() string { return QuestionTypeImage }
	return funcs
}

func Funcs(funcs template.FuncMap, policy *bluemonday.Policy) template.FuncMap {
	if funcs == nil {
		funcs = template.FuncMap{}
	}
	funcs = addQuestionTypes(funcs)
	funcs["idfy"] = idfy
	funcs["FormxMergeQuestionsAnswers"] = MergeQuestionsAnswers
	funcs["FormxAnswerValue"] = answerValue
	funcs["FormxAnswersContainValue"] = answersContainValue
	funcs["FormxSanitizeHTML"] = SanitizeHTML(policy)
	funcs["FormxJoinSlice"] = JoinSlice
	funcs["FormxCheckboxAnswers"] = CheckboxAnswers
	funcs["FormxRadioSelectAnswers"] = RadioSelectAnswers
	funcs["FormxMultiradioAnswers"] = MultiradioAnswers
	return funcs
}

func idfy(inputs ...string) (output string) {
	sha1Bytes := sha1.Sum([]byte(strings.Join(inputs, "")))
	output = hex.EncodeToString(sha1Bytes[:])
	return output[0:8]
}

func MergeQuestionsAnswers(questions Questions, answers Answers) []QuestionAnswer {
	var qas []QuestionAnswer
	for _, question := range questions {
		var qa QuestionAnswer
		qa.Type = question.Type
		qa.Text = question.Text
		qa.Name = question.Name
		qa.Options = append([]Option{}, question.Options...)
		for _, subquestion := range question.Subquestions {
			var subqa SubquestionAnswer
			subqa.Name = subquestion.Name
			subqa.Text = subquestion.Text
			answer := answers[subquestion.Name]
			if len(answer) > 0 {
				subqa.Answer = answer[0]
			}
			qa.SubquestionAnswers = append(qa.SubquestionAnswers, subqa)
		}
		answer := answers[question.Name]
		qa.Answer = append(qa.Answer, answer...)
		qas = append(qas, qa)
	}
	return qas
}

func answerValue(answer []string) string {
	if len(answer) > 0 {
		return answer[0]
	} else {
		return ""
	}
}

func answersContainValue(answer []string, value string) bool {
	for _, a := range answer {
		if a == value {
			return true
		}
	}
	return false
}

func ExtractAnswers(form url.Values, questions []Question) (answers Answers) {
	allAnswers := make(map[string][]string)
	for name, values := range form {
		allAnswers[name] = append(allAnswers[name], values...)
	}
	answers = Answers{}
	for _, qn := range questions {
		switch qn.Type {
		case QuestionTypeMultiradio:
			for _, subqn := range qn.Subquestions {
				answers[subqn.Name] = allAnswers[subqn.Name]
			}
		default:
			answers[qn.Name] = allAnswers[qn.Name]
		}
	}
	return answers
}

func SanitizeHTML(policy *bluemonday.Policy) func(string) template.HTML {
	return func(input string) template.HTML {
		input = strings.ReplaceAll(input, "\\n", "<br>")
		input = policy.Sanitize(input)
		return template.HTML(input)
	}
}

func JoinSlice(slice []string) string {
	return strings.Join(slice, ", ")
}

func CheckboxAnswers(qna QuestionAnswer) string {
	display := make(map[string]string)
	for _, opt := range qna.Options {
		display[opt.Value] = opt.Display
	}
	var displays []string
	for _, value := range qna.Answer {
		displays = append(displays, display[value])
	}
	return JoinSlice(displays)
}

func RadioSelectAnswers(qna QuestionAnswer) string {
	display := make(map[string]string)
	for _, opt := range qna.Options {
		display[opt.Value] = opt.Display
	}
	for _, value := range qna.Answer {
		return display[value]
	}
	return ""
}
func MultiradioAnswers(qna QuestionAnswer, subqna SubquestionAnswer) string {
	display := make(map[string]string)
	for _, opt := range qna.Options {
		display[opt.Value] = opt.Display
	}
	return display[subqna.Answer]
}
