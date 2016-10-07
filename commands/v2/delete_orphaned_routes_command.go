package v2

import (
	"os"

	"code.cloudfoundry.org/cli/actors/v2actions"
	oldCmd "code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands"
	"code.cloudfoundry.org/cli/commands/v2/common"
)

//go:generate counterfeiter . DeleteOrphanedRoutesActor

type DeleteOrphanedRoutesActor interface {
	GetOrphanedRoutes(spaceGUID string) ([]v2actions.Route, v2actions.Warnings, error)
	DeleteRoute(routeGUID string) (v2actions.Warnings, error)
}

type DeleteOrphanedRoutesCommand struct {
	Force           bool        `short:"f" description:"Force deletion without confirmation"`
	usage           interface{} `usage:"CF_NAME delete-orphaned-routes [-f]"`
	relatedCommands interface{} `related_commands:"delete-route, routes"`

	UI     commands.UI
	Actor  DeleteOrphanedRoutesActor
	Config commands.Config
}

func (cmd *DeleteOrphanedRoutesCommand) Setup(config commands.Config, ui commands.UI) error {
	cmd.UI = ui
	cmd.Config = config

	client, err := common.NewCloudControllerClient(config)
	if err != nil {
		return err
	}
	cmd.Actor = v2actions.NewActor(client)

	return nil
}

func (cmd *DeleteOrphanedRoutesCommand) Execute(args []string) error {
	if cmd.Config.Experimental() == false {
		oldCmd.Main(os.Getenv("CF_TRACE"), os.Args)
		return nil
	}

	cmd.UI.DisplayText("This command is in EXPERIMENTAL stage and may change without notice")
	cmd.UI.DisplayNewline()

	err := common.CheckTarget(cmd.Config, true, true)
	if err != nil {
		return err
	}

	if !cmd.Force {
		deleteOrphanedRoutes, err := cmd.UI.DisplayBoolPrompt("Really delete orphaned routes?", false)
		if err != nil {
			return err
		}

		if !deleteOrphanedRoutes {
			return nil
		}
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayHeaderFlavorText("Getting routes as {{.CurrentUser}} ...", map[string]interface{}{
		"CurrentUser": user.Name,
	})

	routes, warnings, err := cmd.Actor.GetOrphanedRoutes(cmd.Config.TargetedSpace().GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	for _, route := range routes {
		cmd.UI.DisplayHeaderFlavorText("Deleting route {{.Route}} ...", map[string]interface{}{
			"Route": route.String(),
		})

		warnings, err = cmd.Actor.DeleteRoute(route.GUID)
		cmd.UI.DisplayWarnings(warnings)
		if err != nil {
			return err
		}
	}

	cmd.UI.DisplayOK()

	return nil
}
