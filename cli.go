package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"time"

	"github.com/codegangsta/cli"
	"github.com/segmentio/go-prompt"
	"github.com/torrick/spyglass/openaperture"
)

func main() {
	app := cli.NewApp()
	environmentFlag := cli.StringFlag{Name: "environment, e", Usage: "environment to build or deploy"}
	commitHashFlag := cli.StringFlag{Name: "commit, c", Usage: "commit hash or branch to build or deploy"}
	app.Name = "spyglass"
	app.Version = "0.2.5"
	app.Author = "Thomas Orrick"
	app.Email = "thomas.orrick@lexmark.com"
	app.Flags = []cli.Flag{
		cli.BoolFlag{Name: "follow, f", Usage: "continually check status of a build or deploy"},
		cli.StringFlag{Name: "server, s", Usage: "build server url", EnvVar: "APERTURE_SERVER_URL"},
	}
	app.Commands = []cli.Command{
		{
			Name:  "deploy",
			Usage: "build and deploy a docker repository",
			Action: func(c *cli.Context) {
				validate(c)
				fmt.Printf("Sending deploy request for:\n Project: %s\n Environment: %s\n", c.Args().First(), c.String("environment"))
				project := deploy(openaperture.NewProject(c.Args().First(), c.String("environment"), c.String("commit"), c.GlobalString("server"),
					c.String("build-exchange"), c.String("deploy-exchange"), c.Bool("force")))
				if c.GlobalBool("follow") {
					checkStatus(project)
				}
			},
			Flags: []cli.Flag{
				environmentFlag,
				commitHashFlag,
				cli.StringFlag{Name: "build-exchange", Usage: "Set the build exchange id"},
				cli.StringFlag{Name: "deploy-exchange", Usage: "Set the deploy exchange id"},
				cli.BoolFlag{Name: "force", Usage: "Force a docker build"},
			},
		},
		{
			Name:  "deploy_ecs",
			Usage: "build and deploy a docker repository",
			Action: func(c *cli.Context) {
				validate(c)
				fmt.Printf("Sending deploy request for:\n Project: %s\n Environment: %s\n", c.Args().First(), c.String("environment"))
				project := deploy_ecs(openaperture.NewProject(c.Args().First(), c.String("environment"), c.String("commit"), c.GlobalString("server"),
					c.String("build-exchange"), c.String("deploy-exchange"), c.Bool("force")))
				if c.GlobalBool("follow") {
					checkStatus(project)
				}
			},
			Flags: []cli.Flag{
				environmentFlag,
				commitHashFlag,
				cli.StringFlag{Name: "build-exchange", Usage: "Set the build exchange id"},
				cli.StringFlag{Name: "deploy-exchange", Usage: "Set the deploy exchange id"},
				cli.BoolFlag{Name: "force", Usage: "Force a docker build"},
			},
		},
		{
			Name:  "configure",
			Usage: "set configuration options",
			Action: func(c *cli.Context) {
				configure(c)
			},
		},
	}
	app.Run(os.Args)
}

func configure(c *cli.Context) {
	username := prompt.String("Username")
	password := prompt.Password("Password")
	config := map[string]string{"username": username, "password": password}
	configJSON, _ := json.Marshal(config)
	ioutil.WriteFile(path.Join(os.Getenv("HOME"), ".aperturecfg"), configJSON, 0600)
}

func checkStatus(project *openaperture.Project) {
	auth, err := openaperture.GetAuth()
	if err != nil {
		panic(err.Error())
	}
	ticker := time.NewTicker(5 * time.Second)
	quit := make(chan struct{})
	for {
		select {
		case <-ticker.C:
			workflow, err := project.Status(auth)
			if err != nil {
				close(quit)
				fmt.Println(err.Error())
				os.Exit(1)
			} else if workflow.WorkflowError {
				close(quit)
				fmt.Println("Workflow failed")
				fmt.Println(workflow.EventLog)
				os.Exit(1)
			} else if workflow.WorkflowCompleted {
				fmt.Printf("Workflow completed in %s\n", workflow.ElapsedWorkflowTime)
				close(quit)
			} else {
				fmt.Printf("Milestone: %s in progress\n", workflow.CurrentStep)
			}
		case <-quit:
			ticker.Stop()
			os.Exit(0)
		}
	}
}

func validate(c *cli.Context) {
	if c.String("environment") == "" {
		fmt.Printf("Environment was not set.  Please specify an environment to %s\n", c.Command.Name)
		os.Exit(1)
	}
	if c.String("commit") == "" {
		fmt.Printf("Commit hash or branch was not set.  Please specify a commit hash or branch to %s\n", c.Command.Name)
		os.Exit(1)
	}
	if c.Args().First() == "" {
		fmt.Printf("Project name was not set. Please specify a project to %s\n", c.Command.Name)
		os.Exit(1)
	}
}

func deploy(project *openaperture.Project) *openaperture.Project {
	operations := []string{"build", "deploy"}
	token, err := openaperture.GetAuth()
	if err != nil {
		panic(err.Error())
	}
	resp, err := project.CreateWorkflow(token, operations)
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("Workflow created: %s\n", resp.Location)
	err = project.ExecuteWorkflow(token, resp.Location)
	if err != nil {
		panic(err.Error())
	}
	fmt.Println("Successfully sent deploy request")
	return project
}

func deploy_ecs(project *openaperture.Project) *openaperture.Project {
	operations := []string{"build", "deploy_ecs"}
	token, err := openaperture.GetAuth()
	if err != nil {
		panic(err.Error())
	}
	resp, err := project.CreateWorkflow(token, operations)
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("Workflow created: %s\n", resp.Location)
	err = project.ExecuteWorkflow(token, resp.Location)
	if err != nil {
		panic(err.Error())
	}
	fmt.Println("Successfully sent deploy request")
	return project
}
