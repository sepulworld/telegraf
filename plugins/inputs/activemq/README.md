# Telegraf plugin: activemq

Get ActiveMQ server stats from XML feed served by the activemq console via HTTP

# Measurements

Meta:

- tags:

  * queue_name 

Measurement names:

- activemq-queues:

  * Tags: `queues`
  * Fields:

    - messages_pending
                - count
    - messages_enqueued
		- count 
    - number_of_consumers
    - number_of_pending_messages

# Example output

Using this configuration:

```
[[inputs.activemq]]
  # Plugin gather metric via parsing XML output of http://127.0.0.1:8161/admin/queues.jsp
  # <property name="authenticate" value="false" /> Needs to be set to false
  #
```

When run with:

```
./telegraf -config telegraf.conf -test -input-filter activemq
```

It produces:

```
Add output here
```

Anyway, just ensure that you can run the command under `telegraf` user, and it
has to produce XML output.
