raftleaderelection
==================
`raftleaderelection` is a golang code - a raft election code which simulates the behaviour of [raft](https://speakerdeck.com/benbjohnson/raft-the-understandable-distributed-consensus-protocol) election protocol. It ensures that only one leader will be present at a time in a cluster. It ensures one leader with consideration of all conditions that may occur.  

#Usage

(note: In the examples below, $ is the shell prompt, and the output of the snippet follows "----------------"
#### Testing the package
```
$ go test
---------------------------------
PASS
ok    Raft   141.193s
```
This command executes test program which contains the test which checks the minory & majority failures. This program kills the servers one by one and checks the only one leader in maximum number of server killing bound.


### Running the packages

```
$ go run main.go 0 11110 /home/rahul/IdeaProjects/cloud/src
--------------------------------
0 in follower
0 in candidate
```
The first parameter is `id`, second one is `port` and third parameter is the `parent directory`
This program runs the program for the above parameters. The configuration for raft are written in `configuration_.xml` files and configuration for cluster package are written in `serverlist_.xml` 
Currently program is written for `5` servers which can be converted into more by modifying the configuration files for above two packages.

# Install

```
go get github.com/RahulWaniIITB/raftleaderelection

```
### The `raftleaderelection/cluster` package

This package contains cluster layer which will be used in making & sending & receiving the messages.

### The `raftleaderelection/Raft` package

This package contains raft leader election layer which will ensure that only one leader at a time. This part simulates the `raft` protocol.
 
# How it works

`raft_test.go` contains the code for testing the software. This program by using `raft` package creates the raft instance and starts its execution. The raft instance created will eventually create the `cluster` instance which will bind network socket with each raft instance.

# License

`raftleaderelection` is available free. So, anyone can use this repository or modify it.
