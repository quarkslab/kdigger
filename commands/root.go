package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/quarkslab/kdigger/pkg/bucket"
	"github.com/quarkslab/kdigger/pkg/plugins/admission"
	"github.com/quarkslab/kdigger/pkg/plugins/apiresources"
	"github.com/quarkslab/kdigger/pkg/plugins/authorization"
	"github.com/quarkslab/kdigger/pkg/plugins/capabilities"
	"github.com/quarkslab/kdigger/pkg/plugins/cgroups"
	"github.com/quarkslab/kdigger/pkg/plugins/devices"
	"github.com/quarkslab/kdigger/pkg/plugins/environment"
	"github.com/quarkslab/kdigger/pkg/plugins/mount"
	"github.com/quarkslab/kdigger/pkg/plugins/node"
	"github.com/quarkslab/kdigger/pkg/plugins/pidnamespace"
	"github.com/quarkslab/kdigger/pkg/plugins/processes"
	"github.com/quarkslab/kdigger/pkg/plugins/runtime"
	"github.com/quarkslab/kdigger/pkg/plugins/services"
	"github.com/quarkslab/kdigger/pkg/plugins/syscalls"
	"github.com/quarkslab/kdigger/pkg/plugins/token"
	"github.com/quarkslab/kdigger/pkg/plugins/userid"
	"github.com/quarkslab/kdigger/pkg/plugins/usernamespace"
	"github.com/quarkslab/kdigger/pkg/plugins/version"
	"github.com/spf13/cobra"
)

// buckets stores all the plugins
var buckets *bucket.Buckets

// var for the output flag
var output string

// var for the output width
var outputWidth int

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
	rootCmd.PersistentFlags().IntVarP(&outputWidth, "width", "w", 140, "Width for the human output")
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
	pidnamespace.Register(buckets)
	usernamespace.Register(buckets)
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
	userid.Register(buckets)
	processes.Register(buckets)
	cgroups.Register(buckets)
	node.Register(buckets)
	apiresources.Register(buckets)
}

// printResults prints results with the output format selected by the flags
func printResults(r bucket.Results, opts bucket.ResultsOpts) error {
	switch output {
	case "human":
		fmt.Print(r.Human(opts))
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

// printError prints error, maybe it would make more sense to return a Results
// struct that can contains the error directly?
func printError(err error, name string) error {
	switch output {
	case "human":
		fmt.Printf("### %s ###\n", strings.ToUpper(name))
		fmt.Printf("Error: %s\n", err.Error())
	case "json":
		jsonErr := struct {
			Bucket string `json:"bucket"`
			Error  string `json:"error"`
		}{
			Bucket: name,
			Error:  err.Error(),
		}

		bJSONErr, err := json.Marshal(jsonErr)
		if err != nil {
			return err
		}
		fmt.Println(string(bJSONErr))
	default:
		return errors.New("internal error, check on output flag must have been done in PersistentPreRunE")
	}
	return nil
}
