package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"syscall"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/spf13/cobra"
)

var rdsConnectCmd = &cobra.Command{
	Use:   "connect [instance-id]",
	Short: "Connect to an RDS instance using psql or mysql",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		instanceID := args[0]
		ctx := context.TODO()
		cfg, err := config.LoadDefaultConfig(ctx)
		if err != nil {
			log.Fatalf("unable to load SDK config, %v", err)
		}

		client := rds.NewFromConfig(cfg)

		output, err := client.DescribeDBInstances(ctx, &rds.DescribeDBInstancesInput{
			DBInstanceIdentifier: &instanceID,
		})
		if err != nil {
			log.Fatalf("failed to describe instance: %v", err)
		}

		if len(output.DBInstances) == 0 {
			log.Fatalf("instance not found: %s", instanceID)
		}

		db := output.DBInstances[0]
		if db.Endpoint == nil {
			log.Fatalf("instance has no endpoint (not available?)")
		}

		host := *db.Endpoint.Address
		port := db.Endpoint.Port
		engine := *db.Engine
		username := "postgres" // Default fallback
		if db.MasterUsername != nil {
			username = *db.MasterUsername
		}

		fmt.Printf("Connecting to %s (%s) at %s:%d...\n", instanceID, engine, host, port)

		var cmdName string
		var cmdArgs []string

		switch engine {
		case "postgres", "aurora-postgresql":
			cmdName = "psql"
			cmdArgs = []string{"-h", host, "-p", fmt.Sprintf("%d", port), "-U", username, "postgres"}
		case "mysql", "mariadb", "aurora-mysql":
			cmdName = "mysql"
			cmdArgs = []string{"-h", host, "-P", fmt.Sprintf("%d", port), "-u", username, "-p"}
		default:
			log.Fatalf("unsupported engine: %s", engine)
		}

		// Look for the binary
		binary, err := exec.LookPath(cmdName)
		if err != nil {
			log.Fatalf("%s not found in PATH. Please install it to connect.", cmdName)
		}

		// Replace process
		env := os.Environ()
		// syscall.Exec requires the first argument to be the command name as well
		execArgs := append([]string{cmdName}, cmdArgs...)
		
		if err := syscall.Exec(binary, execArgs, env); err != nil {
			log.Fatalf("failed to exec: %v", err)
		}
	},
}

func init() {
	rdsCmd.AddCommand(rdsConnectCmd)
}
