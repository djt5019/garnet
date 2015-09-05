------
Garnet
------

A non-blocking Diamond-like metrics collector written in Go.

Why Garnet?
-----------

Unlike other metrics collection systems, Garnet does not serially collect
metrics.  Instead, each collector you want to run does so in it's own
goroutine and will not hold metrics collection up if it takes a
little while to run.  So the metrics that run more frequently will produce
smoother graphs will be smoother and more reliable than those that don't.

Additionally, Garnet uses a "plugin" like system, where the collector is an
executable that writes to a shared Unix socket.  The collectors can be
written in any language that is capable of network communications over a
Unix socket (which is quite a few of them).


Why Build This?
---------------

I'm new to Go and this seems complex enough to be a good learning experience.
Plus, I like metrics and easily pluggable systems.


Why the name Garnet?
--------------------

I picked a random gemstone since Carbon was already taken.


Is this production ready (or even production worthy)?
-----------------------------------------------------

No.


Desired Behavior
----------------

The application itself is set up in such a manner that individual Go executables
will act as collectors.  The main loop will read in a config file which will
indicate a directory containing other config files.  It will read in each one
and determine if a collector for that config will be enabled.  If so, a new
goroutine will be spawned for that collector and will periodically invoke the
command from the config file.  Afterwards, the main loop will hang around dealing
with signals until it is terminated.

The collector goroutine will pass the address to a Unix domain socket as a
command line argument, which the collector application will read in and
connect to.  The collector executable will invoke whatever collection it may
do and will write the results back to Garnet through the Unix socket.

A dedicate aggregation goroutine will read from the Unix socket, format the
data into whatever metric pattern the user has chosen (Graphite,
Metrics 2.0, etc) and send it on it's way.

Main loop
=========

* Read a config file, which includes a directory full of "collector" config files.
* For each collector config file:
    * Add to a list of collectors if ``enabled`` is ``true``.
    * Read the path of the command to execute.
* Set up signal handlers.
    * SIGTERM - Clean up after yourself.
    * SIGHUP  - Reread the config file.
* Create the metrics aggregation goroutine.
* Create a sync.WaitGroup sized to the number of collector entries in the list.
* Create a UNIX socket for IPC (Maybe play with ZMQ, could be fun?)
* For each "collector" spawn a goroutine and pass a reference to the sync.WaitGroup.
* Sleep forever until SIGTERM-ed then clean up.

Collector Goroutines
====================

* Invoke the executable from the config and provide the Unix socket path as an argument
* Sleep for a configurable amount of time before firing again

Metrics Aggregation Goroutine
=============================

* Endlessly consume from the Unix socket
* For each metric payload received
    * Transform the metrics into the desired format from the config
    * Ship the metrics to their intended destination
