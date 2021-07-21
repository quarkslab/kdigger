package commands

import (
	"errors"
	"fmt"
	"os"

	"github.com/mtardy/kdigger/pkg/bucket"
	"github.com/mtardy/kdigger/pkg/plugins/admission"
	"github.com/mtardy/kdigger/pkg/plugins/authorization"
	"github.com/mtardy/kdigger/pkg/plugins/capabilities"
	"github.com/mtardy/kdigger/pkg/plugins/devices"
	"github.com/mtardy/kdigger/pkg/plugins/environment"
	"github.com/mtardy/kdigger/pkg/plugins/mount"
	"github.com/mtardy/kdigger/pkg/plugins/namespaces"
	"github.com/mtardy/kdigger/pkg/plugins/runtime"
	"github.com/mtardy/kdigger/pkg/plugins/services"
	"github.com/mtardy/kdigger/pkg/plugins/syscalls"
	"github.com/mtardy/kdigger/pkg/plugins/token"
	"github.com/mtardy/kdigger/pkg/plugins/version"
	"github.com/spf13/cobra"
)

// buckets stores all the plugins
var buckets *bucket.Buckets

// var for the output flag
var output string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "kdigger",
	Short: "kdigger helps you dig in your Kubernetes cluster",
	Long: `kdigger is an extensible CLI tool to dig around when you are in a Kubernetes
cluster. For that you can use multiples buckets. Buckets are plugins that can
scan specific aspects of a cluster or bring expertise to automate the Kubernetes
pentest process.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if output != "human" && output != "json" {
			return fmt.Errorf("ouput flag must be one of human|json, got %q", output)
		}
		return nil
	},
}

func init() {
	cobra.OnInitialize(registerBuckets)

	rootCmd.PersistentFlags().StringVarP(&output, "output", "o", "human", "Output format. One of: human|json.")
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// registerBuckets registers all the modules into the buckets, newly created
// module should be registered here.
func registerBuckets() {
	buckets = bucket.NewBuckets()

	admission.Register(buckets)
	namespaces.Register(buckets)
	capabilities.Register(buckets)
	environment.Register(buckets)
	token.Register(buckets)
	authorization.Register(buckets)
	syscalls.Register(buckets)
	mount.Register(buckets)
	devices.Register(buckets)
	runtime.Register(buckets)
	services.Register(buckets)
	version.Register(buckets)
}

// printResults prints results with the output format selected by the flags
func printResults(r bucket.Results, opts bucket.ResultsOpts) error {
	switch output {
	case "human":
		fmt.Println(r.Human(opts))
	case "json":
		p, err := r.JSON(opts)
		if err != nil {
			return err
		}
		fmt.Println(p)
	default:
		return errors.New("internal error, check on output flag must have been done in PersistentPreRunE")
	}
	return nil
}
