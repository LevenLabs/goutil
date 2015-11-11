# first-mongo-secondary

When run against an active mongo replica set, will print out deterministically
if the node being run against is the "first" secondary, i.e. the first secondary
to appear in the replica set list of servers.

This is useful for setting up mongo backups, since you don't want all of your
nodes making independant backups. You can have all of your mongo nodes in the
cluster run this, and only one of them will actually get output back.
