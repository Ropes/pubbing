// Copyright © 2016 Josh Roppo joshroppo@gmail.com
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"google.golang.org/cloud"
	"google.golang.org/cloud/pubsub"
)

// pubCmd represents the pub command
var pubCmd = &cobra.Command{
	Use:   "pub",
	Short: "publish messages to defined topic",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Infof("pub called on topic: %s", topic)

		if gceproject == "" || topic == "" {
			log.Errorf("GCE project and topic must be defined")
			os.Exit(1)
		}
		gc := initClient()
		gctx := cloud.NewContext(gceproject, gc)
		log.Infof("gctx: %#v", gctx)

		msg := &pubsub.Message{Data: []byte("hello world")}
		msgIDs, err := pubsub.Publish(gctx, "breckenridge", msg)
		if err != nil {
			log.Errorf("error publishing %v", err)
		}
		log.Infof("message IDs: %#v", msgIDs)

	},
}

func init() {
	RootCmd.AddCommand(pubCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// pubCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// pubCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
