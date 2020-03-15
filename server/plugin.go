package main

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
)

// Plugin implements the interface expected by the Mattermost server to communicate between the server and plugin processes.
type Plugin struct {
	plugin.MattermostPlugin

	// configurationLock synchronizes access to the configuration.
	configurationLock sync.RWMutex

	// configuration is the active plugin configuration. Consult getConfiguration and
	// setConfiguration for usage.
	configuration *configuration
}

func (p *Plugin) OnActivate() error {
	cmd := &model.Command{
		Trigger:          "stress",
		DisplayName:      "Microsoft Calendar",
		Description:      "Interact with your outlook calendar.",
		AutoComplete:     true,
		AutoCompleteDesc: "setup, read, save, teardown, deleteall",
		AutoCompleteHint: "(subcommand)",
	}
	p.API.RegisterCommand(cmd)
	return nil
}

func mErr(err error) *model.AppError {
	return &model.AppError{Message: err.Error()}
}

func (p *Plugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	start := time.Now()

	split := strings.Fields(args.Command)
	command := split[0]
	if command != "/stress" {
		return &model.CommandResponse{
			Text: fmt.Sprintf("%q is not a supported command. I think it's supposed to be /stress", command),
		}, nil
	}

	parameters := []string{}
	subcommand := ""
	if len(split) > 1 {
		subcommand = split[1]
	}
	if len(split) > 2 {
		parameters = split[2:]
	}

	var res string
	switch subcommand {
	case "setup":
		res = p.handleSetup(parameters...)
	case "teardown":
		res = p.handleTeardown(parameters...)
	case "deleteall":
		res = p.handleDeleteAll(parameters...)
	case "read":
		res = p.handleRead(parameters...)
	case "save":
		res = p.handleSave(parameters...)
	case "":
		res = "Index!"
	}

	end := time.Since(start)
	outTime := fmtDuration(end)
	out := fmt.Sprintf("`%s`\n%s\n%s", args.Command, res, outTime)

	return &model.CommandResponse{
		Text: out,
	}, nil
}

func fmtDuration(d time.Duration) string {
	d = d.Round(time.Microsecond)
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second
	d -= s * time.Second
	ms := d / time.Millisecond
	return fmt.Sprintf("%02d min %02d seconds %02d ms", m, s, ms)
}
