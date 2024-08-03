# Distributed Storage

An API-layer service receives REST with a file. Splits the data among X "storage" services.

The same service also can receive a request for getting this file back. Construct it and return.

## Main ideas and assumptions

Chunk Master is the centerpiece of logic. It manages quotas and knows everything about the files.

I think that in general here we want full control via sync communication and ensuring that data is stored on storage services. Communication via queues (and passing chunks via queues) is also an option, and I bet it can work well with proper choreography, but I think orchestration is better here. At least for a quick one-instance version.

This code assumes that all filenames are located in the same flat space, i.e. there is not separation by users, neigher logical nor physical. Auth and separation doesn't seem to be in scope. Also out of scope is duplicated upload handling.

## TODO now

* Premature ending of sending by user - clean up the data
* Storage down detection during uploading/downloading - clean up the data (recovery and faster detection in production version I'm not doing here)
* (?) We don't actually need size or any equal splitting. We can split to chunks as we read the data, sending a new chunk to new storage and then changing storage by asking for a new chunk quota.
* Race conditions on multiple Get/Post/combination
* remoteStorage client - how to pass reader directly without using io.ReadAll?

## What I would do in an Ideal Super Final version

* Auth is missing here. Of course we must have it, especially on read
* Replication && recovery from failure
* Better healthchecks and service discovery
  - grpc.Conn close (now it's a app lifetime connection, but in case of heartbeat detections we should close connections)
* Support for DELETE
* Autoscaling (?)
* Scalable external-facing API service with access via load balancer
* If API services are scalable then I need more elaborate Chunk Master. No concrete thoughts how would I design it for now. Maybe it should be a standalone service. Maybe data still part of API service but with data replication between all of them