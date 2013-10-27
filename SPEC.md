# Monstrics

Monstrics is the metrics monster.


## What's the point

Monitoring by polling is old and ineffective. We have lots of tools sending metrics to graphite, ganglia, statsd and similar tools and should be able to act over those metrics.

Also, watching graphs is tiresome.

Monstrics tries to do something like what [kale](http://codeascraft.com/2013/06/11/introducing-kale/) from etsy does, but on a more alerting/monitoring way.

## Architecture

`metric sources -> AMPQ Exchange -> montrics_server`

![Diagram][1]

Metrics arrive via AMPQ, in graphite format, multiple metrics per message:

```
metric.path.1 value timestamp\n
metric.path.2 value timestamp\n
metric.path.3 value timestamp\n
```

For each 'interesting' metric (the ones defined on the config file to be checked) a new goroutine is launched that does all the heavy work.

In case of an anomaly, the data is sent to another goroutine, waiting only for handling notifications (notification handlers are defined on the config file).

## What are metrics

Metrics are a pair of timestamp and value.

### Metric value types

**Absolute**: the metric represents the value at that exact time

**Rate**: the metric represents a number of occurrences over a given time

## Constraints

Constraints are tests that Monstrics apply to metrics in order to check for troubles.

### Constraints examples (and hopefully to be implemented in the future)

**Limit levels**: A limit that a determined metric can't go under or over. Think disk space percentage or number of available inodes. High water mark, low water mark.

**Rate limit**: A limit on how many occurrences of a value can happen over a determined time. Think errors per second.

**Boredom**: A certain metric has not happened since a determined time. Think keepalive or cron tasks.

**Forecast levels**: A metric is acting out of the forecast value. Think sudden growth of network usage during DDOS attack when compared to previous day.


[1]: http://dl.dropbox.com/s/9z8vrij7ljefu6q/Monstrics%20-%20State%20Diagram.jpeg