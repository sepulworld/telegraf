package activemq 

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

const DefaultUsername = "admin"
const DefaultPassword = "admin"
const DefaultURL = "http://localhost:8161"

type ActiveMQ struct {
	URL      string
	Name     string
	Username string
	Password string
	Queues   []string

	Client *http.Client
}

type Queue  struct {
    Name string `xml:"name,attr"`
    Size string `xml:"size,attr"`
    ConsumerCount string `xml:"consumerCount,attr"`
    EnqueueCount string `xml:"enqueueCount,attr"`
    DequeueCount string `xml:"dequeueCount,attr"`
}

type gatherFunc func(r *ActiveMQ, acc telegraf.Accumulator, errChan chan error)

var gatherFunctions = []gatherFunc{gatherQueues}

var sampleConfig = `
  url = "http://localhost:8161" # required
  # name = "amq-server-1" # optional tag
  # username = "admin"
  # password = "admin"
`

func (r *ActiveMQ) SampleConfig() string {
	return sampleConfig
}

func (r *ActiveMQ) Description() string {
	return "Read metrics from one ActiveMQ server via the admin API"
}

func (r *ActiveMQ) Gather(acc telegraf.Accumulator) error {
	if r.Client == nil {
		r.Client = &http.Client{}
	}

	var errChan = make(chan error, len(gatherFunctions))

	for _, f := range gatherFunctions {
		go f(r, acc, errChan)
	}

	for i := 1; i <= len(gatherFunctions); i++ {
		err := <-errChan
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *ActiveMQ) requestXML(u string, target interface{}) error {
	u = fmt.Sprintf("%s%s", r.URL, u)

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return err
	}

	username := r.Username
	if username == "" {
		username = DefaultUsername
	}

	password := r.Password
	if password == "" {
		password = DefaultPassword
	}

	req.SetBasicAuth(username, password)

	resp, err := r.Client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	xml.NewDecoder(resp.Body).Decode(target)

	return nil
}

func gatherQueues(r *ActiveMQ, acc telegraf.Accumulator, errChan chan error) {
	// Gather information about queues
	queues := make([]Queue, 0)
	err := r.requestJSON("/admin/xml/queues.jsp", &queues)
	if err != nil {
		errChan <- err
		return
	}

	for _, queue := range queues {
		if !r.shouldGatherQueue(queue) {
			continue
		}
		tags := map[string]string{
			"queue":       queue.Name,
		}

		acc.AddFields(
			"activemq_queue",
			map[string]interface{}{
				// common information
				"size":                 queue.Size,
				"consumer_count":       queue.ConsumerCount,
				"enqueue_count":        queue.EnqueueCount,
				"dequeue_count":        queue.DequeueCount,
			},
			tags,
		)
	}

	errChan <- nil
}

func (r *ActiveMQ) shouldGatherQueue(queue Queue) bool {
	if len(r.Queues) == 0 {
		return true
	}

	for _, name := range r.Queues {
		if name == queue.Name {
			return true
		}
	}

	return false
}

func init() {
	inputs.Add("activemq", func() telegraf.Input {
		return &ActiveMQ{}
	})
}
