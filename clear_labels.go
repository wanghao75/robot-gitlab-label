package main

import (
	"fmt"
	"github.com/opensourceways/community-robot-lib/gitlabclient"
	"github.com/xanzy/go-gitlab"
	"strings"

	"k8s.io/apimachinery/pkg/util/sets"
)

func (bot *robot) handleClearLabel(e *gitlab.MergeEvent, cfg *botConfig) error {
	if !gitlabclient.CheckSourceBranchChanged(e) {
		return nil
	}

	labels := sets.NewString()
	for _, v := range e.Labels {
		labels.Insert(v.Name)
	}
	toRemove := getClearLabels(labels, cfg)
	if len(toRemove) == 0 {
		return nil
	}

	number := gitlabclient.GetMRNumber(e)

	if err := bot.cli.RemoveMergeRequestLabel(e.Project.ID, number, toRemove); err != nil {
		return err
	}

	comment := fmt.Sprintf(
		"This pull request source branch has changed, so removes the following label(s): %s.",
		strings.Join(toRemove, ", "),
	)

	return bot.cli.CreateMergeRequestComment(e.Project.ID, number, comment)
}

func getClearLabels(labels sets.String, cfg *botConfig) []string {
	var r []string

	all := labels
	if len(cfg.ClearLabels) > 0 {
		v := all.Intersection(sets.NewString(cfg.ClearLabels...))
		if v.Len() > 0 {
			r = v.UnsortedList()
			all = all.Difference(v)
		}
	}

	exp := cfg.clearLabelsByRegexp
	if exp != nil {
		for k := range all {
			if exp.MatchString(k) {
				r = append(r, k)
			}
		}
	}

	return r
}
