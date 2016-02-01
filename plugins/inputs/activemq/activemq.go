package activemq

import (
	"bytes"
	"encoding/xml"
        "net/http"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"golang.org/x/net/html/charset"
)

const DefaultURL = "http://localhost:8161/admin/xml/queues.jsp"

type activemq struct {
	URL string
        Port string
        Secure bool
        User string
        Password string
}

type queue struct  {
    Name string `xml:"name,attr"`
}

type stats struct {
    Size string `xml:"size,attr"`
    ConsumerCount string `xml:"consumerCount,attr"`
    EnqueueCount string `xml:"enqueueCount,attr"`
    DequeueCount string `xml:"dequeueCount,attr"`
}

var sampleConfig = `
	# Server to gather XML statistics from 
	#
	# If no host is specified, then uses localhost 
	server = "localhost"
        # Admin server port, defaults to 8161
        port   = "8161"
        # Set to true if authentication is required for admin interface, defaults to false
        secure = true
        # Set user and password to auth with, if admin interface is secure
        user = "admin"
        password = "admin"
`

func (r *activemq) SampleConfig() string {
	return sampleConfig
}

func (r *activemq) Description() string {
	return "Read metrics from Activemq admin xml response"
}

func (g *activemq) Gather(acc telegraf.Accumulator) error {
        resp, err := http.Get("http://localhost:8161/admin/xml/queues.jsp")
        resp.Header.Add("Accept", "application/xml")
        resp.Header.Add("Content-Type","application/xml; charset=utf-8")

        if err != nil {
	    return err
        }

        defer resp.Body.Close()

        out, err := ioutil.ReadAll(resp.Body) 

	if err = importMetric(out, acc); err != nil {
		return err
	}

	return nil
}

func importMetric(stat []byte, acc telegraf.Accumulator) error {
	var queue stats

	decoder := xml.NewDecoder(bytes.NewReader(stat))
	decoder.CharsetReader = charset.NewReaderLabel
	if err := decoder.Decode(&p); err != nil {
		return fmt.Errorf("Cannot parse input with error: %v\n", err)
	}

	fields := map[string]interface{}{
                "queue_name":      queue.Name,
		"queue_size":      stats.Size,
		"consumer_count":  stats.ConsumerCount,
		"enqueue_count":   stats.EnqueueCount,
		"dequeue_count":   stats.DequeueCount,
	}
	acc.AddFields("activemq", fields)

	return nil
}

func init() {
	inputs.Add("activemq", func() telegraf.Input {
		return &activemq{}
	})
}
