package skylab

import (
	"bytes"
	"html/template"

	"golang.org/x/net/html"
)

type ValueDisplay struct {
	Value   string
	Display string
}

func InputSelect(vds []ValueDisplay, name, defaultValue, class string) template.HTML {
	node := &html.Node{
		Type: html.ElementNode,
		Data: "select",
		Attr: []html.Attribute{
			{Key: "name", Val: name},
			{Key: "class", Val: class},
		},
	}
	for _, vd := range vds {
		option := &html.Node{
			Type: html.ElementNode,
			Data: "option",
			Attr: []html.Attribute{
				{Key: "value", Val: vd.Value},
			},
			FirstChild: &html.Node{Type: html.TextNode, Data: vd.Display},
		}
		if vd.Value == defaultValue {
			option.Attr = append(option.Attr, html.Attribute{Key: "selected", Val: "selected"})
		}
		node.AppendChild(option)
	}
	buf := &bytes.Buffer{}
	_ = html.Render(buf, node)
	return template.HTML(buf.String())
}

func (skylb Skylab) AddInputSelects(funcs map[string]interface{}) map[string]interface{} {
	if funcs == nil {
		funcs = map[string]interface{}{}
	}
	funcs["SkylabSelectCohort"] = skylb.InputSelectCohort
	funcs["SkylabSelectCohortOptional"] = skylb.InputSelectCohortOptional
	funcs["SkylabSelectStageOptional"] = InputSelectStageOptional
	funcs["SkylabSelectMilestoneOptional"] = InputSelectMilestoneOptional
	funcs["SkylabSelectProjectLevel"] = InputSelectProjectLevel
	funcs["SkylabSelectRole"] = InputSelectRole
	return funcs
}
func (skylb Skylab) InputSelectCohort(name, defaultValue, class string) template.HTML {
	var vds []ValueDisplay
	cohorts := skylb.Cohorts()
	for _, cohort := range cohorts {
		if cohort != "" {
			vds = append(vds, ValueDisplay{Value: cohort, Display: cohort})
		}
	}
	html := InputSelect(vds, name, defaultValue, class)
	return html
}

func (skylb Skylab) InputSelectCohortOptional(name, defaultValue, class string) template.HTML {
	var vds []ValueDisplay
	cohorts := skylb.Cohorts()
	for _, cohort := range cohorts {
		vds = append(vds, ValueDisplay{Value: cohort, Display: cohort})
	}
	html := InputSelect(vds, name, defaultValue, class)
	return html
}

func InputSelectMilestoneOptional(name, defaultValue, class string) template.HTML {
	vds := []ValueDisplay{
		{Value: Milestone1, Display: "Milestone 1"},
		{Value: Milestone2, Display: "Milestone 2"},
		{Value: Milestone3, Display: "Milestone 3"},
		{Value: MilestoneNull, Display: "<No Milestone>"},
	}
	html := InputSelect(vds, name, defaultValue, class)
	return html
}

func InputSelectStageOptional(name, defaultValue, class string) template.HTML {
	vds := []ValueDisplay{
		{Value: StageApplication, Display: "Application"},
		{Value: StageSubmission, Display: "Submission"},
		{Value: StageEvaluation, Display: "Evaluation"},
		{Value: StageFeedback, Display: "Feedback"},
		{Value: StageNull, Display: "<No Stage>"},
	}
	html := InputSelect(vds, name, defaultValue, class)
	return html
}

func InputSelectProjectLevel(name, defaultValue, class string) template.HTML {
	vds := []ValueDisplay{
		{Value: ProjectLevelVostok, Display: "Vostok"},
		{Value: ProjectLevelGemini, Display: "Gemini"},
		{Value: ProjectLevelApollo, Display: "Apollo"},
		{Value: ProjectLevelArtemis, Display: "Artemis"},
	}
	html := InputSelect(vds, name, defaultValue, class)
	return html
}

func InputSelectRole(name, defaultValue, class string) template.HTML {
	vds := []ValueDisplay{
		{Value: RoleStudent, Display: "Student"},
		{Value: RoleAdviser, Display: "Adviser"},
		{Value: RoleMentor, Display: "Mentor"},
		{Value: RoleAdmin, Display: "Admin"},
		{Value: RoleApplicant, Display: "Applicant"},
	}
	html := InputSelect(vds, name, defaultValue, class)
	return html
}
