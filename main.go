package main

import (
	"os"
	"fmt"
	"strconv"
        "strings"
	"net/http"
	_ "crypto/tls"
	"encoding/json"
	
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	
	"github.com/pymhd/go-logging"
	"github.com/pymhd/go-logging/handlers"
)

const (
	gitSecretHeader   = "x-hub-signature"
	BambooUserName    = "bambooUserName"
	BambooPassword    = "BambooPassword"
	BambooURL         = "https://bamboo.com/path/to/api/"
	BambooPlanPostfix = "-RT"
	UI                = "ui/"
)

var (
	labelMap        = make(map[string]bool, 0)
	acceptedActions = map[string]bool{"opened": true, "labeled": true, "reopened": true, "synchronize": true}
	projectMap      = map[string]string{"lynx": "AM", "lynx-ru": "AER", "lynx-in": "AEI", "pymhd/go-simple-cache": "MHD"}
	postfixMap      = map[string]string{"init": "RTIO", "initLite": "RTILTO", "build": "RTBTO", "sonar": "RTSTO", "spring": "RTSCTO", "integration": "RTITO", "uiUnit": "RTUUTO", "unitWeb": "RTUWTO", "update": "RTUTO", "intLite": "RTILO"}
	log             = logger.New("main", handlers.StreamHandler{}, logger.DEBUG, logger.OLEVEL)
	//log             = logger.New("main", handlers.StreamHandler{}, logger.DEBUG, logger.OLEVEL|logger.OTIME|logger.OFILE|logger.OCOLOR)
)

func PostHandler(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	secret, ok := req.Headers[gitSecretHeader]
	if !ok {
		//this request is not from github
		log.Error("Got request without github secret header, exit")
		return genResponse(http.StatusForbidden, http.StatusText(http.StatusForbidden))
	}
	if secret != os.Getenv("SECRET") {
		log.Error("Secert in header does not match secret specifired in aws lambda func")
		return genResponse(http.StatusForbidden, http.StatusText(http.StatusForbidden))
	}
	payload := new(PullRequestPayload)
	if err := json.NewDecoder(strings.NewReader(req.Body)).Decode(payload); err != nil {
		log.Errorf("Could not parse json payload (%s)", err)
		return genResponse(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	log.Debugf("Invoked by action: %s, pull request number: %d, pull request ref: %s, sha: %s by login: %s\n", payload.Action, payload.PullRequest.Number, payload.PullRequest.Head.Ref, payload.PullRequest.Head.Sha, payload.Sender.Login)
	// if action in opened, reopened ...
	if acceptedActions[payload.Action] {
		// check if it labaled, then mark with appropriate taint
		if payload.Action == "labeled" {
			switch payload.Label.Name {
			case "run init test":
				log.Info("labeled action with 'run init test' label")
				labelMap["init"] = true
			case "run init-lite test":
				log.Info("labeled action with 'run-init-lite-test' label")
				labelMap["initLite"] = true
			case "run build test":
				log.Info("labeled action with 'run build test' label")
				labelMap["build"] = true
			case "run sonar test":
				log.Info("labeled action with 'run sonar test' label")
				labelMap["sonar"] = true
			case "run spring test":
				log.Info("labeled action with 'run spring test' label")
				labelMap["spring"] = true
			case "run integration test":
				log.Info("labeled action with 'run integration test' label")
				labelMap["integration"] = true
			case "run ui unit test":
				log.Info("labeled action with 'run ui unit test label' label")
				labelMap["uiUnit"] = true
			case "run unit + web test":
				log.Info("labeled action with 'run unit plus ui test' label")
				labelMap["unitWeb"] = true
			case "run update test":
				log.Info("labeled action with 'run update test' label")
				labelMap["update"] = true
			case "run integration-lite test":
				log.Info("labeled action with 'run integration-lite test' label")
				labelMap["intLite"] = true
			case "RESTARTED":
				if payload.Sender.Login == "user1" || payload.Sender.Login == "user2" {
					log.Info("labeled action with 'restart' label")
					labelMap["restart"] = true
				}
			default:
				log.Warning("Labeled action with unknown label name, skipping...")
				return genResponse(http.StatusOK, "Unsupported label received")
			}
		}
		if err := triggerBambooEndpoint(payload); err != nil {
			log.Error(err)
			return genResponse(http.StatusOK, err.Error())
		}
		return genResponse(http.StatusOK, "bamboo was invoked")
	}
	return genResponse(http.StatusOK, "Unsupported action")
}

func triggerBambooEndpoint(p *PullRequestPayload) error {
	project, ok := projectMap[p.PullRequest.Head.Repo.FullName]
	if !ok {
		log.Warning("Unsupported project")
		return fmt.Errorf("Unsupported repo")
	}
	plan := project + BambooPlanPostfix
	for label, ok := range labelMap {
		if ok {
			plan = fmt.Sprintf("%s-%s", project, postfixMap[label])
			log.Infof("%s test detected, rewriting default plan to: %s \n", label, plan)
		}
	}
	//strange part
	if strings.HasPrefix(p.PullRequest.Head.Ref, UI) {
		if p.Sender.Login == "user1" || p.Sender.Login == "user2" {
			if len(labelMap) == 0 || labelMap["restart"] {
				plan = fmt.Sprintf("%s-%s", project, "RTWMU")
				log.Infof("Case when hed ref starts with 'ui/', special login detected and restart label or no labels set, plan is %s now\n", plan)
			}
		} else {
			plan = fmt.Sprintf("%s-%s", project, "RTU")
			log.Infof("Case when hed ref starts with 'ui/', and no special login detected (labels ignored), plan is %s", plan)

		}
	} else {
		if len(labelMap) == 0 || labelMap["restart"] {
			if p.Sender.Login == "user1" || p.Sender.Login == "user2" {
				plan = fmt.Sprintf("%s-%s", project, "RTWM")
				log.Infof("Case when hed ref does not starts with 'ui/', special login detected and restart label or no labels set, plan is %s now\n", plan)

			}
		}
	}
	makeBambooPostReq(plan, p)
	return nil
}

func genResponse(code int, body string) (events.APIGatewayProxyResponse, error) {
	headers := map[string]string{"content-type": "text/html"}
	response := events.APIGatewayProxyResponse{StatusCode: code, Body: body, Headers: headers}
	return response, nil
}

func makeBambooPostReq(plan string, p *PullRequestPayload) {
	//insecureTransport := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	//client := &http.Client{Transport: insecureTransport}
	pullNum := strconv.Itoa(p.PullRequest.Number)
	req, err := http.NewRequest("POST", BambooURL+plan, nil)
	if err != nil {
		log.Error(err)
	}
	req.SetBasicAuth(BambooUserName, BambooPassword)
	//set params to post request
	q := req.URL.Query()
	q.Add("bamboo.variable.pull_num", pullNum)
	q.Add("bamboo.variable.pull_event", p.Action)
	q.Add("bamboo.variable.sender_login", p.Sender.Login)
	q.Add("bamboo.variable.pull_base_ref", p.PullRequest.Base.Ref)
	q.Add("bamboo.variable.pull_sha", p.PullRequest.Head.Sha)
	req.URL.RawQuery = q.Encode()
	log.Debugf("Bamboo will be triggered: %s\n", req.URL)
	//resp, err := client.Do(req)
	//if err != nil {
	//    log.Error(err)
	//}
	//bodyText, err := ioutil.ReadAll(resp.Body)
	//s := string(bodyText)
	//return s
}

func main() {
	lambda.Start(PostHandler)
}
