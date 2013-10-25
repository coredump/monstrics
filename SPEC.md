# Monstrics

Monstrics is the metrics monster.


## What's the point

Monitoring by polling is old and innefective. We have lots of tools sending metrics to graphite, ganglia, statsd and similar tools and should be able to act over those metrics.

Also, watching graphs is tiresome.

Monstrics tries to do something like what [kale](http://codeascraft.com/2013/06/11/introducing-kale/) from etsy does, but on a more alerting/monitoring way.