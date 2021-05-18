package cmd

import (
	"fmt"

	"github.com/StevenACoffman/j2m"
	"github.com/andygrunwald/go-jira"
	"github.com/cli/cli/pkg/iostreams"
	"github.com/cli/cli/pkg/markdown"
	"github.com/rsteube/carapace"
	"github.com/rsteube/go-jira-cli/cmd/gj/cmd/action"
	"github.com/rsteube/go-jira-cli/internal/api"
	"github.com/rsteube/go-jira-cli/internal/output"
	"github.com/spf13/cobra"
)

var issueViewOpts api.ListIssuesOptions

var issue_viewCmd = &cobra.Command{
	Use:   "view",
	Short: "",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 { // list issues
			issueViewOpts.Fields = []string{"key", "status", "type", "summary", "components", "updated"}
			issues, err := api.ListIssues(cmd.Flag("host").Value.String(), &issueViewOpts)
			if err != nil {
				return err
			}
			return output.Pager(func(io *iostreams.IOStreams) error {
				return output.PrintIssues(io, issues)
			})
		} else { // view issue
			issue, err := api.GetIssue(cmd.Flag("host").Value.String(), args[0], &jira.GetQueryOptions{})
			if err != nil {
				return err
			}

			return output.Pager(func(io *iostreams.IOStreams) error {
				description, err := markdown.Render(j2m.JiraToMD(issue.Fields.Description), "dark")
				if err != nil {
					return err
				}
				fmt.Fprintf(io.Out, "%v %v %v %v\n%v\n", issue.Key, issue.Fields.Status.Name, issue.Fields.Type.Name, issue.Fields.Summary, description)
				return nil
			})
		}
	},
}

func init() {
	issue_viewCmd.Flags().StringSliceVarP(&issueViewOpts.Project, "project", "p", nil, "filter project")
	issue_viewCmd.Flags().StringSliceVarP(&issueViewOpts.Type, "type", "t", nil, "filter type")
	issue_viewCmd.Flags().StringSliceVarP(&issueViewOpts.Status, "status", "s", nil, "filter status")
	issue_viewCmd.Flags().StringSliceVarP(&issueViewOpts.Assignee, "assignee", "a", nil, "filter assignee")
	issue_viewCmd.Flags().StringSliceVarP(&issueViewOpts.Component, "component", "c", nil, "filter component")
	issue_viewCmd.Flags().StringVarP(&issueViewOpts.Query, "query", "q", "", "filter text")
	issueCmd.AddCommand(issue_viewCmd)

	carapace.Gen(issue_viewCmd).FlagCompletion(carapace.ActionMap{
		"component": carapace.ActionMultiParts(",", func(c carapace.Context) carapace.Action {
			return action.ActionComponents(issue_viewCmd, issueViewOpts.Project).Invoke(c).Filter(c.Parts).ToA()
		}),
		"project": carapace.ActionMultiParts(",", func(c carapace.Context) carapace.Action {
			return action.ActionProjects(issue_viewCmd).Invoke(c).Filter(c.Parts).ToA()
		}),
		"status": carapace.ActionMultiParts(",", func(c carapace.Context) carapace.Action {
			return action.ActionStatuses(issue_viewCmd).Invoke(c).Filter(c.Parts).ToA()
		}),
		"type": carapace.ActionMultiParts(",", func(c carapace.Context) carapace.Action {
			return action.ActionIssueTypes(issue_viewCmd, issueViewOpts.Project).Invoke(c).Filter(c.Parts).ToA()
		}),
	})

	carapace.Gen(issue_viewCmd).PositionalCompletion(
		action.ActionIssues(issue_viewCmd, &issueViewOpts),
	)
}
