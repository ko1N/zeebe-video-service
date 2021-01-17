package services

import (
	"context"
	"fmt"

	"github.com/inconshreveable/log15"

	"github.com/zeebe-io/zeebe/clients/go/pkg/entities"
	"github.com/zeebe-io/zeebe/clients/go/pkg/zbc"
)

type Tracker struct {
	client  zbc.Client
	job     entities.Job
	logger  log15.Logger
	logmsgs []string
}

func NewTracker(client zbc.Client, job entities.Job) *Tracker {
	return &Tracker{
		client: client,
		job:    job,
		logger: log15.New("worker", job.GetElementId()),
	}
}

func (self *Tracker) Progress(progress float64) error {
	vars, err := self.client.
		NewSetVariablesCommand().
		ElementInstanceKey(self.job.GetElementInstanceKey()).
		VariablesFromMap(map[string]interface{}{
			fmt.Sprintf("%s_progress", self.job.GetElementId()): fmt.Sprintf("%d%%", int(progress*100.0)),
		})
	if err != nil {
		return err
	}
	vars.Local(false).Send(context.Background())
	return nil
}

func (self *Tracker) log(msg string) error {
	self.logmsgs = append(self.logmsgs, msg)
	vars, err := self.client.
		NewSetVariablesCommand().
		ElementInstanceKey(self.job.GetElementInstanceKey()).
		VariablesFromMap(map[string]interface{}{
			fmt.Sprintf("%s_log", self.job.GetElementId()): self.logmsgs,
		})
	if err != nil {
		return err
	}
	vars.Local(false).Send(context.Background())
	return nil
}

func (self *Tracker) Debug(msg string, ctx ...interface{}) {
	self.logger.Debug(msg, ctx...)
	self.log(fmt.Sprintf("DEBUG: %s", fmt.Sprintf(msg, ctx...)))
}

func (self *Tracker) Info(msg string, ctx ...interface{}) {
	self.logger.Info(msg, ctx...)
	self.log(fmt.Sprintf("INFO: %s", fmt.Sprintf(msg, ctx...)))
}

func (self *Tracker) Warn(msg string, ctx ...interface{}) {
	self.logger.Info(msg, ctx...)
	self.log(fmt.Sprintf("WARN: %s", fmt.Sprintf(msg, ctx...)))
}

func (self *Tracker) Error(msg string, ctx ...interface{}) {
	self.logger.Info(msg, ctx...)
	self.log(fmt.Sprintf("ERROR: %s", fmt.Sprintf(msg, ctx...)))
}

func (self *Tracker) Crit(msg string, ctx ...interface{}) {
	self.logger.Info(msg, ctx...)
	self.log(fmt.Sprintf("CRIT: %s", fmt.Sprintf(msg, ctx...)))
}
