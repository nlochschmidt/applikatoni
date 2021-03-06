package main

import (
	"bytes"
	"encoding/json"

	"log"
	"net/http"
	"text/template"
)

const slackSummaryTmplStr = `{{.GitHubRepo}} {{if .Success}}Successfully Deployed{{else}}Deploy Failed{{end}}:
{{.Username}} deployed {{.Branch}} on {{.Target}} :pizza:

> {{.Comment}}
<{{.GitHubUrl}}|View latest commit on GitHub>
<{{.DeploymentURL}}|Open deployment in Applikatoni>`

var slackTemplate = template.Must(template.New("slackSummary").Parse(slackSummaryTmplStr))

type slackMsg struct {
	Text string `json:"text"`
}

func NotifySlack(ev *DeploymentEvent) {
	if ev.Target.SlackUrl == "" {
		return
	}

	summary, err := generateSummary(slackTemplate, ev)
	if err != nil {
		log.Printf("Could not generate Slack deployment summary, %s\n", err)
		return
	}

	SendSlackRequest(ev, summary)
}

func SendSlackRequest(ev *DeploymentEvent, summary string) {
	payload, err := json.Marshal(slackMsg{Text: summary})

	if err != nil {
		log.Printf("Error creating Slack notification %s\n", err)
		return
	}

	resp, err := http.Post(ev.Target.SlackUrl, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		log.Printf("Notifying Slack failed (%s on %s, %s): err=%s\n",
			ev.Application.Name, ev.Target.Name, ev.Deployment.CommitSha, err)
		return
	}
	if resp.StatusCode != 200 {
		log.Printf("Notifying Slack failed (%s on %s, %s): status=%d\n",
			ev.Application.Name, ev.Target.Name, ev.Deployment.CommitSha,
			resp.StatusCode)
		return
	}

	log.Printf("Successfully notified Slack about deployment of %s on %s, %s!\n",
		ev.Application.Name, ev.Target.Name, ev.Deployment.CommitSha)
}
