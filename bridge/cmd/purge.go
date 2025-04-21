package cmd

import (
	"github.com/spf13/cobra"
	"github.com/streadway/amqp"
	"github.com/zenanetwork/iris/bridge/setu/queue"
	"github.com/zenanetwork/iris/helper"
)

// purgeCmd represents the reset of queue
var purgeCmd = &cobra.Command{
	Use:   "purge-queue",
	Short: "Reset bridge queue tasks",
	Run: func(cmd *cobra.Command, args []string) {
		// purge Queue
		purgeQueue()
	},
}

func purgeQueue() {
	dialer := helper.GetConfig().AmqpURL

	// amqp dialer
	conn, err := amqp.Dial(dialer)
	if err != nil {
		panic(err)
	}

	// initialize exchange
	channel, err := conn.Channel()
	if err != nil {
		panic(err)
	}

	if _, err = channel.QueuePurge(queue.QueueName, false); err != nil {
		logger.Error("purgeQueue | QueuePurge", "Error", err)
	}
}

func init() {
	rootCmd.AddCommand(purgeCmd)
}
