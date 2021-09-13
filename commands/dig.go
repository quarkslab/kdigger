package commands

import (
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

// flag for the color
var color bool

// flag to activate active buckets
var active bool

// flag to force admission creation
var admForce bool

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
			cmd.Help()
		}

		// apply default colored human only if the color flag was not set
		if !cmd.Flags().Changed("color") && output == "human" {
			color = true
		}

		// check if any called buckets are active without the -a flag
		for _, name := range args {
			if buckets.IsActive(name) && !active {
				// Not a good idea finally!
				// this little check is just to facilitate that:
				// "kdigger dig adm -a --adm-force" == "kdigger dig adm --adm-force"
				// canonicalName, _ := buckets.ResolveAlias(name)
				// if canonicalName == "admission" && admForce {
				//     continue
				// }

				return fmt.Errorf("trying to run %q active bucket without the %q or %q flag", name, "--active", "-a")
			}
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// create the config that will be passed to every plugins
		config := &bucket.Config{
			Color:       color,
			OutputWidth: outputWidth,
			AdmForce:    admForce,
		}
		// PreRun guarantee that len(args) != 0 but ...
		if len(args) != 0 {
			// if "all" or "a" just erase all the args with the bucket list
			switch strings.ToLower(args[0]) {
			case "all":
				fallthrough
			case "a":
				if active {
					args = buckets.Registered()
				} else {
					args = buckets.RegisteredPassive()
				}
			}
		}

		// remove duplicates in the args
		alreadyAskBuckets := make(map[string]bool)
		argsWithoutDupl := []string{}
		for _, item := range args {
			if _, found := alreadyAskBuckets[item]; !found {
				alreadyAskBuckets[item] = true
				argsWithoutDupl = append(argsWithoutDupl, item)
			}
		}
		args = argsWithoutDupl

		// iterate through all the specified buckets
		for _, name := range args {
			// all this retry machinery is done to lazy load the client and the
			// checks are in case the plugin return ErrMissingClient forever
			// and we are stuck in an infinite loop. Not the best design...
			retryAttempt := 0
		retryInit:
			if retryAttempt > 1 {
				panic("plugin returns ErrMissingClient after lazy loading the client into the config")
			}
			// intialize the bucket
			b, err := buckets.InitBucket(name, *config)
			if err != nil {
				// config was incomplete for requested bucket
				if err == bucket.ErrMissingClient {
					// lazy load the client, that might seems overkill but it
					// is also really strange to load kubeconfig for buckets
					// that do not need it
					err = loadContext(config)
					if err != nil {
						return err
					}
					retryAttempt++
					goto retryInit
				}
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
	digCmd.Flags().BoolVarP(&active, "active", "a", false, "Enable all buckets that might have side effect on environment.")
	digCmd.Flags().BoolVarP(&admForce, "admission-force", "", false, "Force creation of pods to scan admission even without cleaning rights. (this flag is specific to the admission bucket)")
}
