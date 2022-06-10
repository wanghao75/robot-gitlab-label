package main

import (
	"fmt"
	"github.com/opensourceways/community-robot-lib/gitlabclient"
	"github.com/opensourceways/community-robot-lib/utils"
	"github.com/sirupsen/logrus"
	"github.com/xanzy/go-gitlab"
	"k8s.io/apimachinery/pkg/util/sets"
	"strings"
)

func (bot *robot) handleIssueLabels(
	e *gitlab.IssueCommentEvent,
	toAdd []string,
	toRemove []string,
	cfg *botConfig,
	log *logrus.Entry,
) error {
	lh := genIssueLabelHelper(e, bot.cli)
	if lh == nil {
		return nil
	}

	add := newLabelSet(toAdd)
	remove := newLabelSet(toRemove)
	if v := add.intersection(remove); len(v) > 0 {
		return lh.addComment(fmt.Sprintf(
			"conflict labels(%s) exit", strings.Join(add.origin(v), ", "),
		))
	}

	merr := utils.NewMultiErrors()

	if remove.count() > 0 {
		if _, err := removeLabels(lh, remove); err != nil {
			merr.AddError(err)
		}
	}

	authorID := gitlabclient.GetIssueCommentAuthorID(e)

	if add.count() > 0 {
		err := addLabels(lh, add, authorID, cfg, log)
		if err != nil {
			merr.AddError(err)
		}
	}
	return merr.Err()
}

func genIssueLabelHelper(e *gitlab.IssueCommentEvent, cli iClient) labelHelper {
	rlh := &repoLabelHelper{
		cli: cli,
		pid: e.ProjectID,
	}

	lbs, err := cli.GetIssueLabels(e.ProjectID, e.Issue.IID)
	if err != nil {
		return &issueLabelHelper{
			number:          e.Issue.IID,
			labels:          sets.NewString(),
			repoLabelHelper: rlh,
		}
	}

	return &issueLabelHelper{
		number:          e.Issue.IID,
		labels:          sets.NewString(lbs...),
		repoLabelHelper: rlh,
	}
}
