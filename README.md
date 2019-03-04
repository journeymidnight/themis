# themis

Themis is a tool to dectect vm hosts' failure in openstack environment, and mihgrate virtuail machine using openstack's API.

## Architechture

serf: [https://www.serf.io/docs/internals/gossip.html](https://www.serf.io/docs/internals/gossip.html) is used to detect the 
failure.


Themis is a daemon running on serf. Once serf reports a node failure, and themis will get this messeage and judge by other 
factors to decide whether migrate the hosted virutial machine.

