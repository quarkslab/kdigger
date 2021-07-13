package commands

import (
	"github.com/mtardy/kdigger/pkg/bucket"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:     "ls",
	Aliases: []string{"list"},
	Short:   "List available buckets or describe specific ones",
	Long: `This command lists the available buckets in the binary. It show their names, aliases
and descriptions. You can pass specific buckets as arguments to have their information.`,
	Run: func(cmd *cobra.Command, args []string) {

		var bucketList []string
		if len(args) == 0 {
			bucketList = buckets.Registered()
		} else {
			bucketList = args
		}

		// leveraging bucket results to print even if it's not a plugin
		res := bucket.NewResults("List")
		res.SetHeaders([]string{"Name", "Aliases", "Description"})
		for _, name := range bucketList {
			fullname, found := buckets.ResolveAlias(name)
			if found {
				res.AddContent([]interface{}{fullname, buckets.Aliases(name), buckets.Describe(name)})
			}
		}

		showName := false
		showComment := false
		printResults(*res, bucket.ResultsOpts{ShowName: &showName, ShowComment: &showComment})
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
