package cmd

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/StevenACoffman/j2m"
	"github.com/andygrunwald/go-jira"
	"github.com/cli/browser"
	"github.com/cli/cli/pkg/iostreams"
	"github.com/cli/cli/pkg/markdown"
	"github.com/cli/cli/utils"
	"github.com/rsteube/carapace"
	"github.com/rsteube/go-jira-cli/cmd/gj/cmd/action"
	"github.com/rsteube/go-jira-cli/internal/api"
	"github.com/rsteube/go-jira-cli/internal/output"
	"github.com/spf13/cobra"
)

var issueViewOpts api.ListIssuesOptions

var issue_viewCmd = &cobra.Command{
	Use:   "view",
	Short: "list/view issues",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 { // list issues
			host := cmd.Flag("host").Value.String()
			if cmd.Flag("web").Changed { // open in browser
				jql, err := issueViewOpts.ToJql(host)
				if err != nil {
					return err
				}
				return browser.OpenURL(fmt.Sprintf("https://%v/issues/?jql=%v", host, url.QueryEscape(jql)))
			}

			issueViewOpts.Fields = []string{"key", "status", "type", "summary", "components", "updated"}
			issues, err := api.ListIssues(cmd.Flag("host").Value.String(), &issueViewOpts)
			if err != nil {
				return err
			}
			return output.Pager(func(io *iostreams.IOStreams) error {

				return output.PrintIssues(io, issues)
			})
		} else { // view issue
			if cmd.Flag("web").Changed { // open in browser
				return browser.OpenURL(fmt.Sprintf("https://%v/browse/%v", cmd.Flag("host").Value.String(), args[0]))
			} else {
				issue, err := api.GetIssue(cmd.Flag("host").Value.String(), args[0], &jira.GetQueryOptions{})
				if err != nil {
					return err
				}

				return output.Pager(func(io *iostreams.IOStreams) error {
					description, err := markdown.Render(j2m.JiraToMD(issue.Fields.Description), "dark") // TODO glamour style from config
					if err != nil {
						return err
					}

					components := make([]string, len(issue.Fields.Components))
					for index, component := range issue.Fields.Components {
						components[index] = component.Name
					}

					fmt.Fprintf(io.Out, "%v %v [%v]\n%v %v • opened %v • %v comment(s)\nComponents: %v\nLabels: %v\n%v\n",
						io.ColorScheme().Bold(issue.Key),
						io.ColorScheme().Bold(issue.Fields.Summary),
						io.ColorScheme().Gray(issue.Fields.Priority.Name),
						io.ColorScheme().ColorFromString(issue.Fields.Status.StatusCategory.ColorName)(issue.Fields.Status.Name),
						issue.Fields.Type.Name,
						utils.FuzzyAgo(time.Since(time.Time(issue.Fields.Created))),
						len(issue.Fields.Comments.Comments),
						io.ColorScheme().Gray(strings.Join(components, ", ")),
						io.ColorScheme().Gray(strings.Join(issue.Fields.Labels, ", ")),
						description)
					return nil
				})
			}
		}
	},
}

func init() {
	issue_viewCmd.Flags().Bool("web", false, "view in browser")
	issue_viewCmd.Flags().IntVarP(&issueViewOpts.Filter, "filter", "f", -1, "predefined filter")
	issue_viewCmd.Flags().IntVarP(&issueViewOpts.Limit, "limit", "l", 50, "limit results (default: 50)")
	issue_viewCmd.Flags().StringSliceVar(&issueViewOpts.Label, "label", nil, "filter label")
	issue_viewCmd.Flags().StringSliceVar(&issueViewOpts.Priority, "priority", nil, "filter priority")
	issue_viewCmd.Flags().StringSliceVarP(&issueViewOpts.Assignee, "assignee", "a", nil, "filter assignee")
	issue_viewCmd.Flags().StringSliceVarP(&issueViewOpts.Component, "component", "c", nil, "filter component")
	issue_viewCmd.Flags().StringSliceVarP(&issueViewOpts.Project, "project", "p", nil, "filter project")
	issue_viewCmd.Flags().StringSliceVarP(&issueViewOpts.Resolution, "resolution", "r", nil, "filter resolution")
	issue_viewCmd.Flags().StringSliceVarP(&issueViewOpts.Status, "status", "s", nil, "filter status")
	issue_viewCmd.Flags().StringSliceVar(&issueViewOpts.StatusCategory, "status-category", nil, "filter status-category")
	issue_viewCmd.Flags().StringSliceVarP(&issueViewOpts.Type, "type", "t", nil, "filter type")
	issue_viewCmd.Flags().StringVarP(&issueViewOpts.Jql, "jql", "j", "", "custom jql")
	issue_viewCmd.Flags().StringVarP(&issueViewOpts.Query, "query", "q", "", "filter text")
	issueCmd.AddCommand(issue_viewCmd)

	carapace.Gen(issue_viewCmd).FlagCompletion(carapace.ActionMap{
		"assignee": carapace.ActionMultiParts(",", func(c carapace.Context) carapace.Action {
			return action.ActionUsers(issue_viewCmd).Invoke(c).Filter(c.Parts).ToA() // TODO assignable users
		}),
		"component": carapace.ActionMultiParts(",", func(c carapace.Context) carapace.Action {
			return action.ActionComponents(issue_viewCmd, issueViewOpts.Project).Invoke(c).Filter(c.Parts).ToA()
		}),
		"resolution": carapace.ActionMultiParts(",", func(c carapace.Context) carapace.Action {
			return action.ActionResolutions(issue_viewCmd).Invoke(c).Filter(c.Parts).ToA()
		}),
		"priority": carapace.ActionMultiParts(",", func(c carapace.Context) carapace.Action {
			return action.ActionPriorities(issue_viewCmd).Invoke(c).Filter(c.Parts).ToA()
		}),
		"project": carapace.ActionMultiParts(",", func(c carapace.Context) carapace.Action {
			return action.ActionProjects(issue_viewCmd).Invoke(c).Filter(c.Parts).ToA()
		}),
		"status": carapace.ActionMultiParts(",", func(c carapace.Context) carapace.Action {
			return action.ActionStatuses(issue_viewCmd).Invoke(c).Filter(c.Parts).ToA()
		}),
		"status-category": carapace.ActionMultiParts(",", func(c carapace.Context) carapace.Action {
			return action.ActionStatusCategories(issue_viewCmd).Invoke(c).Filter(c.Parts).ToA()
		}),
		"type": carapace.ActionMultiParts(",", func(c carapace.Context) carapace.Action {
			return action.ActionIssueTypes(issue_viewCmd, issueViewOpts.Project).Invoke(c).Filter(c.Parts).ToA()
		}),
		"filter": action.ActionFilters(issue_viewCmd),
	})

	carapace.Gen(issue_viewCmd).PositionalCompletion(
		action.ActionIssues(issue_viewCmd, &issueViewOpts),
	)
}
