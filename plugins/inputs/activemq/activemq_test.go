package activemq 

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func fakeActiveMQStatus(stat string) {
	content := fmt.Sprintf("#!/bin/sh\ncat << EOF\n%s\nEOF", stat)
	ioutil.WriteFile("/tmp/activemq-status", []byte(content), 0700)
}

func teardown() {
	os.Remove("/tmp/activemq-status")
}

func Test_Invalid_Xml(t *testing.T) {
	fakeActiveMQStatus("invalid xml")
	defer teardown()

	r := &activemq{
		Command: "/tmp/passenger-status",
	}

	var acc testutil.Accumulator

	err := r.Gather(&acc)
	require.Error(t, err)
	assert.Equal(t, err.Error(), "Cannot parse input with error: EOF\n")
}

func TestPassengerGenerateMetric(t *testing.T) {
	fakePassengerStatus(sampleStat)
	defer teardown()

	//Now we tested again above server, with our authentication data
	r := &passenger{
		Command: "/tmp/passenger-status",
	}

	var acc testutil.Accumulator

	err := r.Gather(&acc)
	require.NoError(t, err)

	tags := map[string]string{
		"passenger_version": "5.0.17",
	}
	fields := map[string]interface{}{
		"process_count":      23,
		"max":                23,
		"capacity_used":      23,
		"get_wait_list_size": 3,
	}
	acc.AssertContainsTaggedFields(t, "passenger", fields, tags)

	tags = map[string]string{
		"name":     "/var/app/current/public",
		"app_root": "/var/app/current",
		"app_type": "rack",
	}
	fields = map[string]interface{}{
		"processes_being_spawned": 2,
		"capacity_used":           23,
		"get_wait_list_size":      3,
	}
	acc.AssertContainsTaggedFields(t, "passenger_group", fields, tags)

	tags = map[string]string{
		"name": "/var/app/current/public",
	}

	fields = map[string]interface{}{
		"capacity_used":      23,
		"get_wait_list_size": 3,
	}
	acc.AssertContainsTaggedFields(t, "passenger_supergroup", fields, tags)

	tags = map[string]string{
		"app_root":         "/var/app/current",
		"group_name":       "/var/app/current/public",
		"supergroup_name":  "/var/app/current/public",
		"pid":              "11553",
		"code_revision":    "899ac7f",
		"life_status":      "ALIVE",
		"process_group_id": "13608",
	}
	fields = map[string]interface{}{
		"concurrency":           1,
		"sessions":              0,
		"busyness":              0,
		"processed":             951,
		"spawner_creation_time": int64(1452746835922747),
		"spawn_start_time":      int64(1452746844946982),
		"spawn_end_time":        int64(1452746845013365),
		"last_used":             int64(1452747071764940),
		"uptime":                int64(226), // in seconds of 3m 46s
		"cpu":                   int64(58),
		"rss":                   int64(418548),
		"pss":                   int64(319391),
		"private_dirty":         int64(314900),
		"swap":                  int64(0),
		"real_memory":           int64(314900),
		"vmsize":                int64(1563580),
	}
	acc.AssertContainsTaggedFields(t, "passenger_process", fields, tags)
}

var sampleStat = `
<queues>
  <queue name="prod_email_error">
    <stats size="1364768"
      consumerCount="0"
      enqueueCount="1836607"
      dequeueCount="471839"/>
    <feed>
      <atom>queueBrowse/prod_email_error;jsessionid=ahf00m9tlo0z1hldtmff646nb?view=rss&amp;feedType=atom_1.0</atom>
      <rss>queueBrowse/prod_email_error;jsessionid=ahf00m9tlo0z1hldtmff646nb?view=rss&amp;feedType=rss_2.0</rss>
    </feed>
  </queue>
  <queue name="prod_email_be">
    <stats size="0"
      consumerCount="180"
      enqueueCount="46121234"
      dequeueCount="46121379"/>
    <feed>
      <atom>queueBrowse/prod_email_be;jsessionid=ahf00m9tlo0z1hldtmff646nb?view=rss&amp;feedType=atom_1.0</atom>
      <rss>queueBrowse/prod_email_be;jsessionid=ahf00m9tlo0z1hldtmff646nb?view=rss&amp;feedType=rss_2.0</rss>
    </feed>
  </queue>
</queues>`
