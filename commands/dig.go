package commands

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/quarkslab/kdigger/pkg/automaticontext"
	"github.com/quarkslab/kdigger/pkg/bucket"
	"github.com/spf13/cobra"
	"k8s.io/client-go/util/homedir"
)

// flag for the kubeconfig
var kubeconfig string

// flag for the namespace
var namespace string

// flag to activate side effects buckets
var sideEffects bool

// output formats
const outputHuman = "human"
const outputJSON = "json"

// config that will carry parameters and client for plugin init
var pluginConfig bucket.Config

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
	PreRunE: func(cmd *cobra.Command, args []string) error {
		// display help by default
		if len(args) == 0 {
			return errors.New("missing argument")
		}

		// apply default colored human only if the color flag was not set
		if !cmd.Flags().Changed("color") && output == outputHuman {
			pluginConfig.Color = true
		}

		// check if any called buckets have side effects without the flag activated
		for _, name := range args {
			if buckets.HasSideEffects(name) && !sideEffects {
				// Not a good idea finally!
				// this little check is just to facilitate that:
				// "kdigger dig adm -a --adm-force" == "kdigger dig adm --adm-force"
				// canonicalName, _ := buckets.ResolveAlias(name)
				// if canonicalName == "admission" && admForce {
				//     continue
				// }

				return fmt.Errorf("trying to run %q bucket with side effects without the %q or %q flag", name, "--side-effects", "-s")
			}
		}

		return nil
	},
	RunE: func(_ *cobra.Command, args []string) error {
		// handles the "all" or "a" and erase the args with the bucket list
		// PreRun should guarantee that len(args) != 0 but in case
		if len(args) != 0 {
			switch strings.ToLower(args[0]) {
			case "a":
				fallthrough
			case "all":
				if sideEffects {
					args = buckets.Registered()
				} else {
					args = buckets.RegisteredPassive()
				}
			}
		} else {
			return errors.New("missing argument")
		}

		args = removeDuplicates(args)

		// iterate through all the specified buckets
		// TODO: some plugins might be slow, for example network scanners, so it
		// might be a good idea in the future to parallelize the launch of these
		// plugins
		for _, name := range args {
			// initialize the bucket
			if buckets.RequiresClient(name) {
				err := loadContext(&pluginConfig)
				if err != nil {
					// loading the context failed and is required so skip this
					// execution after printing the error with the name
					err := printError(fmt.Errorf("failed loading context to initialize client: %w", err), name)
					if err != nil {
						return err
					}
					continue
				}
			}
			b, err := buckets.InitBucket(name, pluginConfig)
			if err != nil {
				return err
			}

			// run the bucket
			results, err := b.Run()
			if err != nil {
				err := printError(err, name)
				if err != nil {
					return err
				}
			} else {
				err = printResults(results, bucket.ResultsOpts{OutputWidth: outputWidth})
				if err != nil {
					return err
				}
			}
		}
		return nil
	},
}

func removeDuplicates(list []string) []string {
	set := make(map[string]bool)
	out := []string{}
	for _, item := range list {
		if _, found := set[item]; !found {
			set[item] = true
			out = append(out, item)
		}
	}
	return out
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
	digCmd.Flags().BoolVarP(&sideEffects, "side-effects", "s", false, "Enable all buckets that might have side effect on environment.")

	digCmd.Flags().BoolVarP(&pluginConfig.Color, "color", "c", false, "Enable color in output. (default true if output is human)")
	digCmd.Flags().BoolVarP(&pluginConfig.AdmForce, "admission-force", "", false, "Force creation of pods to scan admission even without cleaning rights. (this flag is specific to the admission bucket)")
	digCmd.Flags().BoolVarP(&pluginConfig.AdmCreate, "admission-create", "", false, "Actually create pods to scan admission instead of using server dry run. (this flag is specific to the admission bucket)")
	// this one is retrieved from the root cmd because applicable to many cmds
	pluginConfig.OutputWidth = outputWidth
}
