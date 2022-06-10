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

func (bot *robot) handleMRLabels(
	e *gitlab.MergeCommentEvent,
	toAdd []string,
	toRemove []string,
	cfg *botConfig,
	log *logrus.Entry,
) error {
	lh := genMRLabelHelper(e, bot.cli)
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

	authorID := gitlabclient.GetMRCommentAuthorID(e)

	if add.count() > 0 {
		err := addLabels(lh, add, authorID, cfg, log)
		if err != nil {
			merr.AddError(err)
		}
	}
	return merr.Err()
}

func genMRLabelHelper(e *gitlab.MergeCommentEvent, cli iClient) labelHelper {
	rlh := &repoLabelHelper{
		cli: cli,
		pid: e.ProjectID,
	}

	lbs, err := cli.GetMergeRequestLabels(e.ProjectID, e.MergeRequest.IID)
	if err != nil {
		return &mrLabelHelper{
			number:          e.MergeRequest.IID,
			labels:          sets.NewString(),
			repoLabelHelper: rlh,
		}
	}

	return &mrLabelHelper{
		number:          e.MergeRequest.IID,
		labels:          sets.NewString(lbs...),
		repoLabelHelper: rlh,
	}
}

func addLabels(lh labelHelper, toAdd *labelSet, commenterID int, cfg *botConfig, log *logrus.Entry) error {
	canAdd, missing, err := checkLabelsToAdd(lh, toAdd, commenterID, cfg, log)
	if err != nil {
		return err
	}

	merr := utils.NewMultiErrors()

	if len(canAdd) > 0 {
		ls := sets.NewString(canAdd...).Difference(lh.getCurrentLabels())
		if ls.Len() > 0 {
			if err := lh.addLabels(ls.UnsortedList()); err != nil {
				merr.AddError(err)
			}
		}
	}

	if len(missing) > 0 {
		msg := fmt.Sprintf(
			"The label(s) `%s` cannot be applied, because the repository doesn't have them",
			strings.Join(missing, ", "),
		)

		if err := lh.addComment(msg); err != nil {
			merr.AddError(err)
		}
	}

	return merr.Err()
}

func checkLabelsToAdd(
	h labelHelper,
	toAdd *labelSet,
	commenterID int,
	cfg *botConfig,
	log *logrus.Entry,
) ([]string, []string, error) {
	v, err := h.getLabelsOfRepo()
	if err != nil {
		return nil, nil, err
	}
	repoLabels := newLabelSet(v)

	missing := toAdd.difference(repoLabels)
	if len(missing) == 0 {
		return repoLabels.origin(toAdd.toList()), nil, nil
	}

	var canAdd []string
	if len(missing) < toAdd.count() {
		canAdd = repoLabels.origin(toAdd.intersection(repoLabels))
	}

	missing = toAdd.origin(missing)

	if !cfg.AllowCreatingLabelsByCollaborator {
		return canAdd, missing, nil
	}

	b, err := h.isCollaborator(commenterID)
	if err != nil {
		return nil, nil, err
	}
	if b {
		if err := h.createLabelsOfRepo(missing); err != nil {
			log.Error(err)
		}

		return append(canAdd, missing...), nil, nil
	}
	return canAdd, missing, nil
}

func removeLabels(lh labelHelper, toRemove *labelSet) ([]string, error) {
	v, err := lh.getLabelsOfRepo()
	if err != nil {
		return nil, err
	}
	repoLabels := newLabelSet(v)

	ls := lh.getCurrentLabels().Intersection(sets.NewString(
		repoLabels.origin(toRemove.toList())...)).UnsortedList()

	if len(ls) == 0 {
		return nil, nil
	}
	return ls, lh.removeLabels(ls)
}
