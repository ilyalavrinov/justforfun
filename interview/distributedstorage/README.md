# Distributed Storage

An API-layer service receives REST with a file. Splits the data among X "storage" services.

The same service also can receive a request for getting this file back. Construct it and return.

## Main ideas

Chunk Master is the centerpiece of logic. It manages quotas and knows everything about the files.

I think that in general here we want full control via sync communication and ensuring that data is stored on storage services. Communication via queues (and passing chunks via queues) is also an option, and I bet it can work well with proper choreography, but I think orchestration is better here. At least for a quick one-instance version.

## TODO now

* Premature ending of sending by user - clean up the data
* Storage down detection during uploading/downloading - clean up the data (recovery and faster detection in production version I'm not doing here)

## What I would do in an Ideal Super Final version

* Auth is missing here. Of course we must have it, especially on read
* Replication && recovery from failure
* Better healthchecks and service discovery
* Support for DELETE
* Autoscaling (?)
* Scalable external-facing API service with access via load balancer
* Multiple API services - more elaborate Chunk Master. No concrete thoughts how would I design it for now