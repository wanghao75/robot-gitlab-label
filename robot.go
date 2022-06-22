package main

import (
	"github.com/opensourceways/community-robot-lib/gitlabclient"
	"github.com/xanzy/go-gitlab"

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

func newRobot(cli iClient, gc func() (*configuration, error)) *robot {
	return &robot{getConfig: gc, cli: cli}
}

type robot struct {
	getConfig func() (*configuration, error)
	cli       iClient
}

func (bot *robot) HandleMergeEvent(e *gitlab.MergeEvent, log *logrus.Entry) error {
	org, repo := gitlabclient.GetMROrgAndRepo(e)

	c, err := bot.getConfig()
	if err != nil {
		return err
	}
	botCfg := c.configFor(org, repo)

	merr := utils.NewMultiErrors()
	if err = bot.handleClearLabel(e, botCfg); err != nil {
		merr.AddError(err)
	}

	mrID := gitlabclient.GetMRNumber(e)
	projectID := e.Project.ID

	commits, err := bot.cli.GetMergeRequestCommits(projectID, mrID)
	if err != nil {
		merr.AddError(err)
	}

	commitsCount := len(commits)

	if err = bot.handleSquashLabel(e, commitsCount, projectID, botCfg.SquashConfig); err != nil {
		merr.AddError(err)
	}

	return merr.Err()
}

func (bot *robot) HandleMergeCommentEvent(e *gitlab.MergeCommentEvent, log *logrus.Entry) error {
	if e.ObjectKind != "note" || e.MergeRequest.State != "opened" {
		log.Debug("Event is not a creation of a comment or MR is not opened, skipping.")
		return nil
	}

	org, repo := gitlabclient.GetMRCommentOrgAndRepo(e)
	c, err := bot.getConfig()
	if err != nil {
		return err
	}
	botCfg := c.configFor(org, repo)

	toAdd, toRemove := getMatchedLabels(gitlabclient.GetMRCommentBody(e))
	if len(toAdd) == 0 && len(toRemove) == 0 {
		log.Debug("invalid comment, skipping.")
		return nil
	}

	return bot.handleMRLabels(e, toAdd, toRemove, botCfg, log)
}

func (bot *robot) HandleIssueCommentEvent(e *gitlab.IssueCommentEvent, log *logrus.Entry) error {
	if e.ObjectKind != "note" || e.Issue.State != "opened" {
		log.Debug("Event is not a creation of a comment or MR is not opened, skipping.")
		return nil
	}

	org, repo := gitlabclient.GetIssueCommentOrgAndRepo(e)
	c, err := bot.getConfig()
	if err != nil {
		return err
	}
	botCfg := c.configFor(org, repo)

	toAdd, toRemove := getMatchedLabels(gitlabclient.GetIssueCommentBody(e))
	if len(toAdd) == 0 && len(toRemove) == 0 {
		log.Debug("invalid comment, skipping.")
		return nil
	}

	return bot.handleIssueLabels(e, toAdd, toRemove, botCfg, log)
}
