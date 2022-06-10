package main

import (
	"regexp"

	"github.com/opensourceways/community-robot-lib/config"
)

type configuration struct {
	ConfigItems []botConfig `json:"config_items,omitempty"`
}

func (c *configuration) configFor(org, repo string) *botConfig {
	if c == nil {
		return nil
	}

	items := c.ConfigItems
	v := make([]config.IRepoFilter, len(items))

	for i := range items {
		v[i] = &items[i]
	}

	if i := config.Find(org, repo, v); i >= 0 {
		return &items[i]
	}

	return nil
}

func (c *configuration) Validate() error {
	if c == nil {
		return nil
	}

	items := c.ConfigItems
	for i := range items {
		if err := items[i].validate(); err != nil {
			return err
		}
	}

	return nil
}

func (c *configuration) SetDefault() {
	if c == nil {
		return
	}

	Items := c.ConfigItems
	for i := range Items {
		Items[i].setDefault()
	}
}

type botConfig struct {
	config.RepoFilter

	// ClearLabels specifies labels that should be removed when the codes of PR are changed.
	ClearLabels []string `json:"clear_labels,omitempty"`

	// ClearLabelsByRegexp specifies a expression which can match a list of labels that
	// should be removed when the codes of PR are changed.
	ClearLabelsByRegexp string `json:"clear_labels_by_regexp,omitempty"`
	clearLabelsByRegexp *regexp.Regexp

	// AllowCreatingLabelsByCollaborator is a tag which will lead to create unavailable labels
	// by collaborator if it is true.
	AllowCreatingLabelsByCollaborator bool `json:"allow_creating_labels_by_collaborator,omitempty"`

	SquashConfig
}

func (c *botConfig) setDefault() {
	c.SquashConfig.setDefault()
}

func (c *botConfig) validate() error {
	if c.ClearLabelsByRegexp != "" {
		v, err := regexp.Compile(c.ClearLabelsByRegexp)
		if err != nil {
			return err
		}
		c.clearLabelsByRegexp = v
	}

	return c.RepoFilter.Validate()
}

type SquashConfig struct {
	// UnableCheckingSquash indicates whether unable checking squash.
	UnableCheckingSquash bool `json:"unable_checking_squash,omitempty"`

	// CommitsThreshold Check the threshold of the number of PR commits,
	// and add the label specified by SquashCommitLabel to the PR if this value is exceeded.
	// zero means no check.
	CommitsThreshold int `json:"commits_threshold,omitempty"`

	// SquashCommitLabel Specify the label whose PR exceeds the threshold. default: stat/needs-squash
	SquashCommitLabel string `json:"squash_commit_label,omitempty"`
}

func (c *SquashConfig) setDefault() {
	if c.CommitsThreshold == 0 {
		c.CommitsThreshold = 1
	}

	if c.SquashCommitLabel == "" {
		c.SquashCommitLabel = "stat/needs-squash"
	}
}

func (c SquashConfig) unableCheckingSquash() bool {
	return c.UnableCheckingSquash
}
