package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func main() {
	// Load AWS SDK configuration
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		fmt.Printf("Unable to load SDK config: %v\n", err)
		return
	}

	// Create EC2 client
	ec2Client := ec2.NewFromConfig(cfg)

	// Create tview application
	app := tview.NewApplication()

	// Create table
	table := tview.NewTable().
		SetBorders(true).
		SetFixed(1, 0)

	// Set table headers
	headers := []string{"Instance ID", "Name", "Type", "State", "Private IP", "Public IP", "AZ", "Instance Lifecycle"}
	for i, header := range headers {
		table.SetCell(0, i,
			tview.NewTableCell("  "+header+"  "). // Add spaces for padding
				SetTextColor(tcell.ColorYellow).
				SetSelectable(false).
				SetAlign(tview.AlignCenter).
				SetAttributes(tcell.AttrBold).
				SetExpansion(1))
	}

	// Get EC2 instances
	result, err := ec2Client.DescribeInstances(context.TODO(), &ec2.DescribeInstancesInput{})
	if err != nil {
		fmt.Printf("Unable to describe instances: %v\n", err)
		return
	}

	// Populate table with instance data
	row := 1
	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {
			// Get instance name from tags
			var instanceName string
			for _, tag := range instance.Tags {
				if *tag.Key == "Name" {
					instanceName = *tag.Value
					break
				}
			}

			// Get IP addresses (handling nil pointers)
			privateIP := ""
			if instance.PrivateIpAddress != nil {
				privateIP = *instance.PrivateIpAddress
			}
			publicIP := ""
			if instance.PublicIpAddress != nil {
				publicIP = *instance.PublicIpAddress
			}

			// Determine instance lifecycle
			instanceLifecycle := "On-Demand"
			if instance.InstanceLifecycle != types.InstanceLifecycleType("") {
				instanceLifecycle = string(instance.InstanceLifecycle)
			}

			// Add row to table
			cols := []string{
				*instance.InstanceId,
				instanceName,
				string(instance.InstanceType),
				string(instance.State.Name),
				privateIP,
				publicIP,
				*instance.Placement.AvailabilityZone,
				instanceLifecycle,
			}

			for col, content := range cols {
				table.SetCell(row, col,
					tview.NewTableCell("  "+content+"  "). // Add spaces for padding
						SetTextColor(tcell.ColorWhite).
						SetAlign(tview.AlignCenter).
						SetExpansion(1))
			}
			row++
		}
	}

	// Set up key bindings
	table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape || event.Rune() == 'q' {
			app.Stop()
		}
		return event
	})

	// Set table properties
	table.SetSelectable(false, false) // Disable selecting rows using arrow keys

	// Run application
	if err := app.SetRoot(table, true).EnableMouse(true).Run(); err != nil {
		fmt.Printf("Error running application: %v\n", err)
	}
}