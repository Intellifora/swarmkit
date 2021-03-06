package task

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/docker/swarmkit/api"
	"github.com/docker/swarmkit/cmd/swarmctl/common"
	"github.com/docker/swarmkit/protobuf/ptypes"
	"github.com/spf13/cobra"
)

func printTaskStatus(w io.Writer, t *api.Task) {
	fmt.Fprintln(w, "Status\t")
	fmt.Fprintf(w, "  Desired State\t: %s\n", t.DesiredState.String())
	fmt.Fprintf(w, "  Last State\t: %s\n", t.Status.State.String())
	if t.Status.Timestamp != nil {
		fmt.Fprintf(w, "  Timestamp\t: %s\n", ptypes.TimestampString(t.Status.Timestamp))
	}
	if t.Status.Message != "" {
		fmt.Fprintf(w, "  Message\t: %s\n", t.Status.Message)
	}
	if t.Status.Err != "" {
		fmt.Fprintf(w, "  Error\t: %s\n", t.Status.Err)
	}
	ctnr := t.Status.GetContainer()
	if ctnr == nil {
		return
	}
	if ctnr.ContainerID != "" {
		fmt.Fprintf(w, "  ContainerID:\t: %s\n", ctnr.ContainerID)
	}
	if ctnr.PID != 0 {
		fmt.Fprintf(w, "  Pid\t: %d\n", ctnr.PID)
	}
	if t.Status.State > api.TaskStateRunning {
		fmt.Fprintf(w, "  ExitCode\t: %d\n", ctnr.ExitCode)
	}
}

func printTaskSummary(task *api.Task, res *common.Resolver) {
	w := tabwriter.NewWriter(os.Stdout, 8, 8, 8, ' ', 0)
	defer w.Flush()

	fmt.Fprintf(w, "ID\t: %s\n", task.ID)
	fmt.Fprintf(w, "Instance\t: %d\n", task.Instance)
	fmt.Fprintf(w, "Service\t: %s\n", res.Resolve(api.Service{}, task.ServiceID))
	printTaskStatus(w, task)
	fmt.Fprintf(w, "Node\t: %s\n", res.Resolve(api.Node{}, task.NodeID))

	fmt.Fprintln(w, "Spec\t")
	ctr := task.Spec.GetContainer()
	common.FprintfIfNotEmpty(w, "  Image\t: %s\n", ctr.Image)
	common.FprintfIfNotEmpty(w, "  Command\t: %q\n", strings.Join(ctr.Command, " "))
	common.FprintfIfNotEmpty(w, "  Args\t: [%s]\n", strings.Join(ctr.Args, ", "))
	common.FprintfIfNotEmpty(w, "  Env\t: [%s]\n", strings.Join(ctr.Env, ", "))
}

var (
	inspectCmd = &cobra.Command{
		Use:   "inspect <task ID>",
		Short: "Inspect a task",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("task ID missing")
			}
			c, err := common.Dial(cmd)
			if err != nil {
				return err
			}

			t, err := c.GetTask(common.Context(cmd), &api.GetTaskRequest{TaskID: args[0]})
			if err != nil {
				return err
			}
			task := t.Task

			// TODO(aluzzardi): This should be implemented as a ListOptions filter.
			r, err := c.ListTasks(common.Context(cmd), &api.ListTasksRequest{})
			if err != nil {
				return err
			}
			previous := []*api.Task{}
			for _, t := range r.Tasks {
				if t.ServiceID == task.ServiceID && t.Instance == task.Instance {
					previous = append(previous, t)
				}
			}

			res := common.NewResolver(cmd, c)

			printTaskSummary(task, res)
			if len(previous) > 0 {
				fmt.Printf("\n===> Instance History\n")
				Print(previous, true, res)
			}

			return nil
		},
	}
)
