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
	"io/ioutil"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/cloud"
	"google.golang.org/cloud/pubsub"
)

var (
	subscription string
	numConsume   int
	quit         chan os.Signal
)

// shouldQuit listens on the quit channel and returns true
// if the signal has been caught and closes the channel.
func shouldQuit(quit chan os.Signal) bool {
	select {
	case q := <-quit:
		log.Warnf("quit signal sent: %v", q)
		signal.Stop(quit)
		close(quit)
		return true
	default:
		log.Debugf("shouldQuit defaulting")
		return false
	}
}

// JWTClientInit reads in a service account JSON token and creates an oauth
// token for communicating with GCE.
func JWTClientInit(ctx *context.Context) *pubsub.Client {
	jsonKey, err := ioutil.ReadFile(keyPath)
	if err != nil {
		log.Errorf("error reading keyfile: %v", err)
		os.Exit(1)
	}

	conf, err := google.JWTConfigFromJSON(jsonKey, pubsub.ScopePubSub)
	if err != nil {
		log.Errorf("error creating conf file: %v", err)
	}

	oauthTokenSource := conf.TokenSource(*ctx)
	psClient, err := pubsub.NewClient(*ctx, gceproject, cloud.WithTokenSource(oauthTokenSource))
	if err != nil {
		log.Errorf("error creating pubsub client: %v", err)
		os.Exit(1)
	}
	return psClient
}

// GCEClientInit uses Google's host FS searching functionality to find auth
// tokens if they exist. eg: GCE VMs, Authenticated Developers
func GCEClientInit(ctx *context.Context, project string) *pubsub.Client {
	var client *pubsub.Client
	clientOnce := new(sync.Once)
	clientOnce.Do(func() {
		source, err := google.DefaultTokenSource(*ctx, pubsub.ScopePubSub)
		if err != nil {
			log.Errorf("error creating token source: %v", err)
			os.Exit(1)
		}
		client, err = pubsub.NewClient(*ctx, project, cloud.WithTokenSource(source))
		if err != nil {
			log.Errorf("error creating pubsub.Client: %v", err)
			os.Exit(1)
		}
	})
	return client
}

// subCmd represents the sub command
var subCmd = &cobra.Command{
	Use:   "sub",
	Short: "subscribe to messages",
	Long:  `Subscribe to messages from a specified topic and subscription.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Infof("sub called on topic: %s", topic)

		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

		if gceproject == "" || topic == "" || subscription == "" {
			log.Errorf("GCE project, subscription, and topic must be defined")
			os.Exit(1)
		}

		ctx := context.Background()
		var psClient *pubsub.Client

		if keyPath != "" {
			psClient = JWTClientInit(&ctx)
		} else {
			psClient = GCEClientInit(&ctx, gceproject)
		}
		if psClient == nil {
			log.Errorf("PubSub client is nil")
			os.Exit(1)
		}

		log.Debugf("client: %#v", psClient)
		sub := psClient.Subscription(subscription)

		it, err := sub.Pull(ctx, pubsub.MaxExtension(time.Second*5))
		if err != nil {
			log.Errorf("error creating pubsub iterator: %v", err)
		}
		defer it.Stop()

		for !shouldQuit(quit) {
			m, err := it.Next()
			if err != nil {
				log.Errorf("error reading from iterator: %v", err)
			}
			log.Infof("msg[%s] read: %v", m.ID, m.Data)
		}

		os.Exit(0)
	},
}

func init() {
	RootCmd.AddCommand(subCmd)
	RootCmd.PersistentFlags().StringVar(&subscription, "sub", "", "PubSub subscription")
	RootCmd.PersistentFlags().IntVar(&numConsume, "num", 10, "Messages to consume")
}
