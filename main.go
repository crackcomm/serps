package main

import (
	"encoding/json"
	"os"

	"github.com/golang/glog"
	"gopkg.in/urfave/cli.v2"

	clinsq "github.com/crackcomm/cli-nsq"
	"github.com/crackcomm/nsqueue/consumer"
	"github.com/crackcomm/nsqueue/producer"
	"github.com/crackcomm/serps/search"

	r "gopkg.in/gorethink/gorethink.v3"
)

var flags = []cli.Flag{
	clinsq.TopicFlag,
	clinsq.ChannelFlag,
	clinsq.AddrFlag,
	clinsq.LookupAddrFlag,
	&cli.StringFlag{
		Name:    "rethink-addr",
		EnvVars: []string{"RETHINK_ADDR"},
		Usage:   "rethinkdb address",
		Value:   "rethink:28015",
	},
	&cli.StringFlag{
		Name:    "rethink-db",
		EnvVars: []string{"RETHINK_DB"},
		Usage:   "rethinkdb database name",
		Value:   "default",
	},
	&cli.StringFlag{
		Name:    "rethink-table",
		EnvVars: []string{"RETHINK_TABLE"},
		Usage:   "rethinkdb table name",
		Value:   "serps",
	},
	&cli.StringSliceFlag{
		Name:    "crawl-topic",
		EnvVars: []string{"CRAWL_TOPIC"},
		Usage:   "nsq search results topic",
	},
	&cli.StringSliceFlag{
		Name:    "crawl-callback",
		EnvVars: []string{"CRAWL_CALLBACK"},
		Usage:   "search results crawl callback",
	},
}

type crawlRequest struct {
	URL       string   `json:"url,omitempty"`
	Referer   string   `json:"referer,omitempty"`
	Callbacks []string `json:"callbacks,omitempty"`
}

func main() {
	defer glog.Flush()

	app := (&cli.App{})
	app.Name = "serps"
	app.Usage = "serp store"
	app.Flags = flags
	app.Before = clinsq.RequireAll
	app.Action = appMain

	if err := app.Run(os.Args); err != nil {
		glog.Fatal(err)
	}
}

func appMain(c *cli.Context) (err error) {
	// Connect to rethink database
	session, err := r.Connect(r.ConnectOpts{
		Address: c.String("rethink-addr"),
	})
	if err != nil {
		return
	}
	// Defer closing database connection
	defer func() {
		if err := session.Close(); err != nil {
			glog.Error(err)
		}
	}()

	// Get rethinkdb table pointer
	table := r.DB(c.String("rethink-db")).Table(c.String("rethink-table"))

	resultHandler := func(msg *consumer.Message) {
		// Read search result from message body
		var result search.Result
		if err := msg.ReadJSON(&result); err != nil {
			glog.Errorf("Read JSON error: %v", err)
			msg.GiveUp()
			return
		}

		// Construct crawl requests from URLs
		var reqs []interface{}
		for _, url := range result.Results {
			reqs = append(reqs, crawlRequest{
				URL:       url,
				Referer:   result.Source,
				Callbacks: c.StringSlice("crawl-callback"),
			})
		}

		// Marshal all requests to JSON
		bodies, err := multiJSONMarshal(reqs)
		if err != nil {
			glog.Errorf("Requests JSON marshal error: %v", err)
			msg.GiveUp()
			return
		}

		// Publish crawl requests to all topics
		for _, topic := range c.StringSlice("crawl-topic") {
			if err := producer.MultiPublish(topic, bodies); err != nil {
				glog.Errorf("Publish error: %v", err)
				msg.Fail()
				return
			}
		}

		// Create and set result ID
		result.ID = search.GetID(result)

		// Store search results in a database
		if err := table.Insert(result).Exec(session); err != nil {
			return
		}

		// Message processing is done
		msg.Success()
		return
	}

	// Register consumer of search results
	for _, topic := range c.StringSlice("topic") {
		consumer.Register(topic, c.String("channel"), 1, resultHandler)
	}

	// Connect after registering consumers
	err = clinsq.Connect(c)
	if err != nil {
		return
	}

	// Start consuming search results from NSQ
	consumer.Start(false)
	return
}

func multiJSONMarshal(values []interface{}) (_ [][]byte, err error) {
	res := make([][]byte, len(values))
	for n, v := range values {
		body, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		res[n] = body
	}
	return
}
