package commands

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/mtardy/kdigger/pkg/automaticontext"
	"github.com/mtardy/kdigger/pkg/bucket"
	"github.com/spf13/cobra"
	"k8s.io/client-go/util/homedir"
)

// var for the kubeconfig flag
var kubeconfig string

// var for the namespace flag
var namespace string

// var for the color flag
var color bool

// digCmd represents the dig command
var digCmd = &cobra.Command{
	Use:     "dig [buckets]",
	Aliases: []string{"d"},
	Short:   "Use all buckets or specific ones",
	Long: `This command runs buckets, special keyword "all" or "a" runs all registered
buckets. You can find information about all buckets with the list command. To
run one or more specific buckets, just input their names or aliases as
arguments.`,
	// This two lines will not work because buckets.Registered() because
	// buckets will be nil at evaluation
	// ValidArgs: buckets.Registered(),
	// Args:      cobra.OnlyValidArgs,
	PreRun: func(cmd *cobra.Command, args []string) {
		// apply default colored human only if the color flag was not set
		if !cmd.Flags().Changed("color") && output == "human" {
			color = true
		}

		if len(args) == 0 {
			cmd.Help()
		}
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// create the config that will be passed to every plugins
		config := bucket.NewConfig()
		config.Color = color
		config.OutputWidth = outputWidth

		// PreRun guarantee that len(args) != 0 but ...
		if len(args) != 0 {
			// if "all" or "a" just erase all the args with the bucket list
			switch strings.ToLower(args[0]) {
			case "all":
				fallthrough
			case "a":
				args = buckets.Registered()
			}
		}

		// iterate through all the specified buckets
		for _, name := range args {
		retry:
			b, err := buckets.InitBucket(name, *config)

			if err != nil {
				// config was incomplete for requested bucket
				if err == bucket.ErrMissingClient {
					// lazy load the client, that might seems overkill but it
					// is also really strange to load kubeconfig for buckets
					// that do not need it
					err = loadContext(config)
					if err != nil {
						panic(err)
					}
					// we have a potential infinite loop if the plugin return
					// ErrMissingClient forever... not a great design
					goto retry
				}
				return err
			}

			results, err := b.Run()
			if err != nil {
				panic(err)
			}

			err = printResults(results, bucket.ResultsOpts{OutputWidth: outputWidth})
			if err != nil {
				panic(err)
			}
		}
		return nil
	},
}

// loadContext loads the kubernetes client and the current namespace into the
// config
func loadContext(config *bucket.Config) error {
	// if we load the client, we will also need the current namespace
	if namespace != "" {
		config.Namespace = namespace
	} else {
		ns, err := automaticontext.CurrentNamespace()
		if err != nil {
			return err
		}
		config.Namespace = ns
	}

	cf, err := automaticontext.Client(kubeconfig)
	if err != nil {
		return err
	}
	config.Client = cf

	return nil
}

func init() {
	rootCmd.AddCommand(digCmd)

	// kubeconfig flags
	if home := homedir.HomeDir(); home != "" {
		path := filepath.Join(home, ".kube", "config")
		_, err := os.Stat(path)
		if !os.IsNotExist(err) {
			digCmd.Flags().StringVar(&kubeconfig, "kubeconfig", path, "(optional) absolute path to the kubeconfig file")
		}
	} else {
		digCmd.Flags().StringVar(&kubeconfig, "kubeconfig", "", "absolute path to the kubeconfig file")
	}

	digCmd.Flags().StringVarP(&namespace, "namespace", "n", "", "Kubernetes namespace to use. (default to the namespace in the context)")
	digCmd.Flags().BoolVarP(&color, "color", "c", false, "Enable color in output. (default true if output is human)")
}
