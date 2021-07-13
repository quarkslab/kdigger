package commands

import (
	"log"
	"os"
	"path/filepath"

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
	Use:     "dig",
	Aliases: []string{"d"},
	Short:   "Use all buckets or specific ones",
	Long: `This command, with no arguments, runs all registered buckets. You can find
information about all buckets with the list command. To run one or more
specific buckets, just input their names or aliases as arguments.`,
	// TODO you can create a validation function for args with cobra
	// This two lines will not work because buckets.Registered() is empty at
	// the beginning
	// ValidArgs: buckets.Registered(),
	// Args: cobra.OnlyValidArgs,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if !cmd.Flags().Changed("color") && output == "human" {
			color = true
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		// create the config that will be passed to every plugins
		config := bucket.NewConfig()
		config.Color = color

		// if no args provided, use all buckets
		if len(args) == 0 {
			args = buckets.Registered()
		}

		for _, name := range args {
		retry:
			b, err := buckets.InitBucket(name, *config)

			if err != nil {
				// config was incomplete for requested bucket
				if err == bucket.ErrMissingClient {
					// lazy load the client
					// that might seems overkill but it is really strange
					// to load kubeconfig for buckets that do not need it
					err = loadContext(config)
					if err != nil {
						panic(err)
					}
					// this is bullshit because we have a potential infinite
					// loop if the plugin return ErrMissingClient forever
					// TODO!!!!
					goto retry
				}
				if _, ok := err.(bucket.ErrUnknownBucket); ok {
					log.Fatal(err)
				}
				panic(err)
			}

			results, err := b.Run()
			if err != nil {
				panic(err)
			}

			err = printResults(results, bucket.ResultsOpts{})
			if err != nil {
				panic(err)
			}
		}
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

func fileExist(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
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
