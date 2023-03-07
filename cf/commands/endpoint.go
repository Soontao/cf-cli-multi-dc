package commands

import (
	"fmt"
	"strings"

	"code.cloudfoundry.org/cli/cf/api/organizations"
	"code.cloudfoundry.org/cli/cf/api/spaces"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/flags"
	. "code.cloudfoundry.org/cli/cf/i18n"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/terminal"
)

type Endpoint struct {
	ui        terminal.UI
	config    coreconfig.ReadWriter
	orgRepo   organizations.OrganizationRepository
	spaceRepo spaces.SpaceRepository
}

func init() {
	commandregistry.Register(&Endpoint{})
}

func (cmd *Endpoint) MetaData() commandregistry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["a"] = &flags.StringFlag{ShortName: "a", Usage: T("api endpoint pattern")}

	return commandregistry.CommandMetadata{
		Name:        "endpoint",
		ShortName:   "e",
		Description: T("Set cf cli endpoint"),
		Usage: []string{
			T("CF_NAME e -a [API Endpoint pattern]"),
		},
		Flags: fs,
	}
}

func (cmd *Endpoint) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	usageReq := requirements.NewUsageRequirement(commandregistry.CLICommandUsagePresenter(cmd),
		T("No argument required"),
		func() bool {
			return len(fc.Args()) != 0
		},
	)

	reqs := []requirements.Requirement{
		usageReq,
		requirementsFactory.NewAPIEndpointRequirement(),
	}

	if !fc.IsSet("a") {
		return nil, fmt.Errorf("Incorrect usage: api endpoint is required", len(fc.Args()), 1)
	}
	return reqs, nil
}

func (cmd *Endpoint) SetDependency(deps commandregistry.Dependency, _ bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.orgRepo = deps.RepoLocator.GetOrganizationRepository()
	cmd.spaceRepo = deps.RepoLocator.GetSpaceRepository()
	return cmd
}

func (cmd *Endpoint) Execute(c flags.FlagContext) error {
	apiEndpointPattern := c.String("a")

	for _, i := range cmd.config.InstanceData() {
		if strings.Contains(i.AuthorizationEndpoint, apiEndpointPattern) {
			cmd.ui.Say("Found existed endpoint %s", i.AuthorizationEndpoint)
			return nil
		}
	}

	err := cmd.ui.ShowConfiguration(cmd.config)
	if err != nil {
		return err
	}
	cmd.ui.NotifyUpdateIfNeeded(cmd.config)
	if !cmd.config.IsLoggedIn() {
		return fmt.Errorf("") // Done on purpose, do not redo this in refactor code
	}
	return nil
}
