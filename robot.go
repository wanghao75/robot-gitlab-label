package main

import (
	"fmt"
	"github.com/opensourceways/community-robot-lib/gitlabclient"
	"github.com/xanzy/go-gitlab"

	"github.com/opensourceways/community-robot-lib/config"
	framework "github.com/opensourceways/community-robot-lib/robot-gitlab-framework"
	"github.com/opensourceways/community-robot-lib/utils"
	"github.com/sirupsen/logrus"
)

const (
	botName    = "label"
	actionOpen = "open"
)

type iClient interface {
	RemoveMergeRequestLabel(projectID interface{}, mrID int, labels gitlab.Labels) error
	CreateMergeRequestComment(projectID interface{}, mrID int, comment string) error
	GetMergeRequestCommits(projectID interface{}, mrID int) ([]*gitlab.Commit, error)
	AddMergeRequestLabel(projectID interface{}, mrID int, labels gitlab.Labels) error
	IsCollaborator(projectID interface{}, loginID int) (bool, error)
	RemoveIssueLabels(projectID interface{}, issueID int, labels gitlab.Labels) error
	AddIssueLabels(projectID interface{}, issueID int, labels gitlab.Labels) error
	CreateIssueComment(projectID interface{}, issueID int, comment string) error
	GetProjectLabels(projectID interface{}) ([]*gitlab.Label, error)
	CreateProjectLabel(pid interface{}, label, color string) error
	GetMergeRequestLabels(projectID interface{}, mrID int) (gitlab.Labels, error)
	GetIssueLabels(projectID interface{}, issueID int) ([]string, error)
}

func newRobot(cli iClient) *robot {
	return &robot{cli: cli}
}

type robot struct {
	cli iClient
}

func (bot *robot) NewConfig() config.Config {
	return &configuration{}
}

func (bot *robot) RobotName() string {
	return botName
}

func (bot *robot) getConfig(cfg config.Config, org, repo string) (*botConfig, error) {
	c, ok := cfg.(*configuration)
	if !ok {
		return nil, fmt.Errorf("can't convert to configuration")
	}

	if bc := c.configFor(org, repo); bc != nil {
		return bc, nil
	}

	return nil, fmt.Errorf("no config for this repo:%s/%s", org, repo)
}

func (bot *robot) RegisterEventHandler(p framework.HandlerRegister) {
	p.RegisterIssueCommentHandler(bot.handleIssueCommentEvent)
	p.RegisterMergeCommentEventHandler(bot.handleMergeCommentEvent)
	p.RegisterMergeEventHandler(bot.handleMREvent)
}

func (bot *robot) handleMREvent(e *gitlab.MergeEvent, pc config.Config, log *logrus.Entry) error {
	org, repo := gitlabclient.GetMROrgAndRepo(e)

	cfg, err := bot.getConfig(pc, org, repo)
	if err != nil {
		return err
	}

	merr := utils.NewMultiErrors()
	if err = bot.handleClearLabel(e, cfg); err != nil {
		merr.AddError(err)
	}

	mrID := gitlabclient.GetMRNumber(e)
	projectID := e.Project.ID

	commits, err := bot.cli.GetMergeRequestCommits(projectID, mrID)
	if err != nil {
		merr.AddError(err)
	}

	commitsCount := len(commits)

	if err = bot.handleSquashLabel(e, commitsCount, projectID, cfg.SquashConfig); err != nil {
		merr.AddError(err)
	}

	return merr.Err()
}

func (bot *robot) handleMergeCommentEvent(e *gitlab.MergeCommentEvent, pc config.Config, log *logrus.Entry) error {
	if e.ObjectKind != "note" || e.MergeRequest.State != "opened" {
		log.Debug("Event is not a creation of a comment or MR is not opened, skipping.")
		return nil
	}

	org, repo := gitlabclient.GetMRCommentOrgAndRepo(e)
	cfg, err := bot.getConfig(pc, org, repo)
	if err != nil {
		return err
	}

	toAdd, toRemove := getMatchedLabels(gitlabclient.GetMRCommentBody(e))
	if len(toAdd) == 0 && len(toRemove) == 0 {
		log.Debug("invalid comment, skipping.")
		return nil
	}

	return bot.handleMRLabels(e, toAdd, toRemove, cfg, log)
}

func (bot *robot) handleIssueCommentEvent(e *gitlab.IssueCommentEvent, pc config.Config, log *logrus.Entry) error {
	if e.ObjectKind != "note" || e.Issue.State != "opened" {
		log.Debug("Event is not a creation of a comment or MR is not opened, skipping.")
		return nil
	}

	org, repo := gitlabclient.GetIssueCommentOrgAndRepo(e)
	cfg, err := bot.getConfig(pc, org, repo)
	if err != nil {
		return err
	}

	toAdd, toRemove := getMatchedLabels(gitlabclient.GetIssueCommentBody(e))
	if len(toAdd) == 0 && len(toRemove) == 0 {
		log.Debug("invalid comment, skipping.")
		return nil
	}

	return bot.handleIssueLabels(e, toAdd, toRemove, cfg, log)
}
