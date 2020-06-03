package v7

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type DownloadDropletCommand struct {
	BaseCommand

	RequiredArgs    flag.AppName `positional-args:"yes"`
	Droplet         string       `long:"droplet" description:"The guid of the droplet to download (default: app's current droplet)."`
	usage           interface{}  `usage:"CF_NAME download-droplet APP_NAME [--droplet DROPLET_GUID]"`
	relatedCommands interface{}  `related_commands:"apps, droplets, push, set-droplet"`

	// field for setting current working dir for ease of testing
	CWD string
}

func (cmd DownloadDropletCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	var (
		rawDropletBytes []byte
		dropletGUID     string
		warnings        v7action.Warnings
	)
	if cmd.Droplet != "" {
		dropletGUID = cmd.Droplet
		cmd.UI.DisplayTextWithFlavor("Downloading droplet {{.DropletGUID}} for app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...", map[string]interface{}{
			"DropletGUID": dropletGUID,
			"AppName":     cmd.RequiredArgs.AppName,
			"OrgName":     cmd.Config.TargetedOrganization().Name,
			"SpaceName":   cmd.Config.TargetedSpace().Name,
			"Username":    user.Name,
		})

		rawDropletBytes, warnings, err = cmd.Actor.DownloadDropletByGUIDAndAppName(dropletGUID, cmd.RequiredArgs.AppName, cmd.Config.TargetedSpace().GUID)
		cmd.UI.DisplayWarnings(warnings)
	} else {
		cmd.UI.DisplayTextWithFlavor("Downloading current droplet for app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...", map[string]interface{}{
			"AppName":   cmd.RequiredArgs.AppName,
			"OrgName":   cmd.Config.TargetedOrganization().Name,
			"SpaceName": cmd.Config.TargetedSpace().Name,
			"Username":  user.Name,
		})

		rawDropletBytes, dropletGUID, warnings, err = cmd.Actor.DownloadCurrentDropletByAppName(cmd.RequiredArgs.AppName, cmd.Config.TargetedSpace().GUID)
		cmd.UI.DisplayWarnings(warnings)
	}
	if err != nil {
		if _, ok := err.(actionerror.DropletNotFoundError); ok {
			return translatableerror.NoDropletForAppError{AppName: cmd.RequiredArgs.AppName, DropletGUID: cmd.Droplet}
		}
		return err
	}

	var pathToDroplet string
	if cmd.CWD == "" {
		currentDir, err := os.Getwd()
		if err != nil {
			return err
		}
		cmd.CWD = currentDir
	}
	pathToDroplet = filepath.Join(cmd.CWD, fmt.Sprintf("droplet_%s.tgz", dropletGUID))

	err = ioutil.WriteFile(pathToDroplet, rawDropletBytes, 0666)
	if err != nil {
		return translatableerror.DropletFileError{Err: err}
	}

	cmd.UI.DisplayText("Droplet downloaded successfully at {{.FilePath}}", map[string]interface{}{
		"FilePath": pathToDroplet,
	})
	cmd.UI.DisplayOK()

	return nil
}
