package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"text/tabwriter"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/spf13/cobra"
)

var rdsCmd = &cobra.Command{
	Use:   "rds",
	Short: "Check Amazon RDS instance status",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.TODO()
		cfg, err := config.LoadDefaultConfig(ctx)
		if err != nil {
			log.Fatalf("unable to load SDK config, %v", err)
		}

		client := rds.NewFromConfig(cfg)

		output, err := client.DescribeDBInstances(ctx, &rds.DescribeDBInstancesInput{})
		if err != nil {
			log.Fatalf("failed to describe instances, %v", err)
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tSTATUS\tENGINE\tENDPOINT")
		
		for _, db := range output.DBInstances {
			endpoint := ""
			if db.Endpoint != nil {
				endpoint = fmt.Sprintf("%s:%d", *db.Endpoint.Address, db.Endpoint.Port)
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", *db.DBInstanceIdentifier, *db.DBInstanceStatus, *db.Engine, endpoint)
		}
		w.Flush()
	},
}

func init() {
	rootCmd.AddCommand(rdsCmd)
}
