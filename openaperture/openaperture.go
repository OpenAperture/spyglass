package openaperture

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// Project Struct
type Project struct {
	Name             string
	Environment      string
	Commit           string
	Server           string
	WorkflowID       string
	ForceBuild       bool
	BuildExchangeID  string
	DeployExchangeID string
}

// Workflow Struct
type Workflow struct {
	EventLog              []string          `json:"event_log"`
	WorkflowCompleted     bool              `json:"workflow_completed"`
	WorkflowError         bool              `json:"workflow_error"`
	CreatedAt             string            `json:"created_at"`
	CurrentStep           string            `json:"current_step"`
	DeploymentRepo        string            `json:"deployment_repo"`
	DeploymentRepoGitRef  string            `json:"deployment_repo_git_ref"`
	ElapsedStepTime       string            `json:"elapsed_step_time"`
	ElapsedWorkflowTime   string            `json:"elapsed_workflow_time"`
	ID                    string            `json:"id"`
	Milestones            []string          `json:"milestones"`
	SourceCommitHash      string            `json:"source_commit_hash"`
	SourceRepo            string            `json:"source_repo"`
	SourceRepoGitRef      string            `json:"source_repo_git_ref"`
	UpdatedAt             string            `json:"updated_at"`
	WorkflowDuration      string            `json:"workflow_duration"`
	WorkflowStepDurations map[string]string `json:"workflow_step_durations"`
}

//Request struct
type Request struct {
	DeploymentRepo   string   `json:"deployment_repo"`
	DeploymentGitRef string   `json:"deployment_repo_git_ref"`
	SourceRepo       string   `json:"source_repo"`
	SourceRepoGitRef string   `json:"source_repo_git_ref"`
	Milestones       []string `json:"milestones"`
}

//ExecuteRequest struct
type ExecuteRequest struct {
	BuildExchangeID  string `json:"build_messaging_exchange_id,omitempty"`
	DeployExchangeID string `json:"deploy_messaging_exchange_id, omitempty"`
	ForceBuild       bool   `json:"force_build"`
}

//ApertureResponse struct
type ApertureResponse struct {
	Status     string
	StatusCode int
	Location   string
	Body       []byte
}

// NewProject Builds a Project object
func NewProject(projectName string, environment string, commit string, server string, buildExchangeID string,
	deployExchangeID string, forceBuild bool) *Project {
	return &Project{Name: projectName, Environment: environment, Commit: commit, Server: server,
		BuildExchangeID: buildExchangeID, DeployExchangeID: deployExchangeID, ForceBuild: forceBuild}
}

// NewRequest builds a new OpenAperture request
func (project *Project) NewRequest(operations []string) *Request {
	return &Request{DeploymentRepo: project.Name, DeploymentGitRef: project.Environment,
		SourceRepoGitRef: project.Commit, Milestones: operations}
}

// NewExecuteRequest builds a new ExecuteRequest
func (project *Project) NewExecuteRequest(forceBuild bool) *ExecuteRequest {
	return &ExecuteRequest{BuildExchangeID: project.BuildExchangeID, DeployExchangeID: project.DeployExchangeID, ForceBuild: forceBuild}
}

// CreateWorkflow initalizes the specifed project workflow
func (project *Project) CreateWorkflow(auth *Auth, operations []string) (*ApertureResponse, error) {
	uri := fmt.Sprintf("%s/workflows", project.Server)
	payload, _ := json.Marshal(project.NewRequest(operations))
	resp, err := httpRequest("POST", auth, uri, payload)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusCreated {
		return nil, errors.New("Failed to create workflow: " + resp.Status)
	}
	project.WorkflowID = strings.Split(resp.Location, "/")[2]
	return resp, nil
}

// ExecuteWorkflow sends the request to OpenAperture to execute the specified workflow
func (project *Project) ExecuteWorkflow(auth *Auth, workflowURI string) error {
	uri := fmt.Sprintf("%s%s/execute", project.Server, workflowURI)
	payload, _ := json.Marshal(project.NewExecuteRequest(project.ForceBuild))
	resp, err := httpRequest("POST", auth, uri, payload)
	if err != nil {
		return err
	}
	if resp.StatusCode == http.StatusNoContent || resp.StatusCode == http.StatusAccepted {
		return nil
	}
	return errors.New("Failed to execute workflow: " + resp.Status)
}

// Status checks the status of a workflow
func (project *Project) Status(auth *Auth) (*Workflow, error) {
	var workflow *Workflow
	path := fmt.Sprintf("%s/workflows/%s", project.Server, project.WorkflowID)
	resp, err := httpRequest("GET", auth, path, nil)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("Failed to retrieve status: " + resp.Status)
	}
	json.Unmarshal(resp.Body, &workflow)
	return workflow, nil
}

func httpRequest(method string, auth *Auth, uri string, payload []byte) (*ApertureResponse, error) {
	var location string
	req, _ := http.NewRequest(method, uri, bytes.NewBuffer(payload))
	req.Header.Set("Authorization", auth.GetAuthorizationHeader())
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	if resp.Header["Location"] != nil {
		location = resp.Header["Location"][0]
	}
	return &ApertureResponse{Status: resp.Status, StatusCode: resp.StatusCode, Body: body, Location: location}, nil
}
