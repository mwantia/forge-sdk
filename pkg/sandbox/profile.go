package sandbox

import (
	"context"
	"os/exec"
)

var ExecProfiles = []ExecProfile{
	{
		Name:        "bash",
		Extensions:  []string{"sh"},
		Arguments:   []string{},
		Environment: []string{},
	},
	{
		Name:        "python3",
		Extensions:  []string{"py"},
		Arguments:   []string{},
		Environment: []string{},
	},
	{
		Name:        "node",
		Extensions:  []string{"js"},
		Arguments:   []string{},
		Environment: []string{},
	},
}

type ExecProfile struct {
	Name        string
	Extensions  []string
	Arguments   []string
	Environment []string
}

func (p *ExecProfile) Cmd(ctx context.Context, arguments []string, environment []string) *exec.Cmd {
	arg := append(p.Arguments, arguments...)
	env := append(p.Environment, environment...)

	cmd := exec.CommandContext(ctx, p.Name, arg...)
	cmd.Env = env

	return cmd
}
