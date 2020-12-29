package models

// Replica is the exported type for which replica to set to/get from
type Replica int

// MainReplica is the variable used for the follower partition
var MainReplica Replica = 0

// BackupReplica is the variable used for the follower partition
var BackupReplica Replica = 1
