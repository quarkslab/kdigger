package commands

import (
	"os"

	"github.com/quarkslab/kdigger/pkg/kgen"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/printers"
)

var opts kgen.GenerateOpts

var genAll bool

var genCmd = &cobra.Command{
	Use:     "gen [name] [flags]",
	Aliases: []string{"generate"},
	Short:   "Generate template for pod with security features disabled",
	Long: `This command generates templates for pod with security features disabled.
You can customize the pods with some of the string flags and activate
boolean flags to disabled security features. Examples:

  # Generate a very simple template in json
  kdigger gen -o json

  # Create a very simple pod
  kdigger gen | kubectl apply -f -

  # Create a pod named mypod with most security features disabled
  kdigger gen -all mypod | kubectl apply -f -
  
  # Create a custom privileged pod
  kdigger gen --privileged --image bash --command watch --command date | kubectl apply -f -`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// all puts all the boolean flags to true
		if genAll {
			opts.HostNetwork = true
			opts.Privileged = true
			opts.HostPath = true
			opts.HostPid = true
			opts.Tolerations = true
		}
		if len(args) > 0 {
			opts.Name = args[0]
		}

		pod := kgen.Generate(opts)

		var p printers.ResourcePrinter
		if output == outputJSON {
			p = &printers.JSONPrinter{}
		} else {
			p = &printers.YAMLPrinter{}
		}
		err := p.PrintObj(pod, os.Stdout)
		if err != nil {
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(genCmd)
	genCmd.Flags().StringVarP(&opts.Namespace, "namespace", "n", "", "Kubernetes namespace to use")
	genCmd.Flags().StringVar(&opts.Image, "image", "busybox", "Container image used")
	genCmd.Flags().StringArrayVar(&opts.Command, "command", []string{"sleep", "infinitely"}, "Container command used")

	genCmd.Flags().BoolVar(&opts.Privileged, "privileged", false, "Add the security flag to the security context of the pod")
	genCmd.Flags().BoolVar(&opts.HostPath, "hostpath", false, "Add a hostPath volume to the container")
	genCmd.Flags().BoolVar(&opts.Tolerations, "tolerations", false, "Add tolerations to be schedulable on most nodes")
	genCmd.Flags().BoolVar(&opts.HostPid, "hostpid", false, "Add the hostPid flag on the whole pod")
	genCmd.Flags().BoolVar(&opts.HostNetwork, "hostnetwork", false, "Add the hostNetwork flag on the whole pod")

	genCmd.Flags().BoolVar(&genAll, "all", false, "Enable everything")
}
