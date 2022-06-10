package main

import (
	"github.com/opensourceways/community-robot-lib/gitlabclient"
	"github.com/xanzy/go-gitlab"
	"k8s.io/apimachinery/pkg/util/sets"
)

func (bot *robot) handleSquashLabel(e *gitlab.MergeEvent, commits, pid int, cfg SquashConfig) error {
	if cfg.unableCheckingSquash() {
		return nil
	}

	action := e.ObjectAttributes.Action
	if action != actionOpen && !gitlabclient.CheckSourceBranchChanged(e) {
		return nil
	}

	labels := sets.NewString()
	for _, v := range e.Labels {
		labels.Insert(v.Name)
	}
	hasSquashLabel := labels.Has(cfg.SquashCommitLabel)
	exceeded := commits > cfg.CommitsThreshold
	number := gitlabclient.GetMRNumber(e)

	if exceeded && !hasSquashLabel {
		return bot.cli.AddMergeRequestLabel(pid, number, gitlab.Labels{cfg.SquashCommitLabel})
	}

	if !exceeded && hasSquashLabel {
		return bot.cli.RemoveMergeRequestLabel(pid, number, gitlab.Labels{cfg.SquashCommitLabel})
	}

	return nil
}
