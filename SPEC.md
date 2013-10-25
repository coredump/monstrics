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

[1]: http://f.cl.ly/items/1Q113G1a122C1a470L47/Screen%20Shot%202013-10-25%20at%204.55.16%20PM.png