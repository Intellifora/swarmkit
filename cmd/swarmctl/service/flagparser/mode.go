package flagparser

import (
	"fmt"

	"github.com/docker/swarmkit/api"
	"github.com/spf13/pflag"
)

func parseMode(flags *pflag.FlagSet, spec *api.ServiceSpec) error {
	if flags.Changed("mode") {
		mode, err := flags.GetString("mode")
		if err != nil {
			return err
		}

		switch mode {
		case "global":
			if spec.GetGlobal() == nil {
				spec.Mode = &api.ServiceSpec_Global{
					Global: &api.GlobalService{},
				}
			}
		case "replicated":
			if spec.GetReplicated() == nil {
				spec.Mode = &api.ServiceSpec_Replicated{
					Replicated: &api.ReplicatedService{},
				}
			}
		}
	}

	if flags.Changed("instances") {
		if spec.GetReplicated() == nil {
			return fmt.Errorf("--instances can only be specified in --mode replicated")
		}
		instances, err := flags.GetUint64("instances")
		if err != nil {
			return err
		}
		spec.GetReplicated().Instances = instances
	}

	return nil
}
