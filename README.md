# vigil

_A file monitor that runs a command on file changes_

I originally named the WIP `vigilo` from the latin, meaning
`to watch through, spend in watching, do or make while watching`, but then I
found [vigilo](https://github.com/wycats/vigilo) from [wycats](https://github.com/wycats)
that does virtually the same thing. I changed the name to `vigil` based on the
name of one of my favorite [Lamb of God songs](https://www.youtube.com/watch?v=lxgelwqe8-E).

## Why?
My google foo was lacking when I was looking for a thing to run tests for a
golang project, watch for file changes, stop an existig test run, and re-run
the tests continuously. There's a section lower with links to a bunch of
existing, likely more hardened, projects below.

## Why poll?
It's less efficient than using a file system notification API, but doesn't run
into problems with the maximum number of open files per process.

## Why proto and grpc?
I was interviewing and it was mentioned the company's backend was golang/proto/grpc
based. I did plenty of proto/grpc at Google, but not using Go, and I've found
the best way to learn about a less familiar tech is to write a slightly more
than trivial project using that tech.

## Generate the proto stuff

1. Install `protoc`
1. Install `protoc-gen-go`
1. Install `protoc-gen-go-grpc`
1. From the root of the vigil project directory
    `protoc --go_out=proto/ --go-grpc_out=proto/ --proto_path=proto/ proto/vigil.proto`

## How it works
By default the main entry point to the program starts a server and a watcher
that maintains a map of file paths from the currenty working directory ->
modified time and if the modified time changes, cancels any running instance
of the given command and starts a new instance.

The main entrypoint also provides a flag to run the program as a client, which
can either make a request to add a path to the list of watched files, or
alternatively return the list of watched files without adding any new paths.
If a directory is passed, it is watched recursively.

## Similar Projects
* [https://github.com/guard/guard](https://github.com/guard/guard)
* [https://github.com/alexch/rerun](https://github.com/alexch/rerun)
* [https://github.com/mynyml/watchr](https://github.com/mynyml/watchr)
* [https://github.com/eaburns/Watch](https://github.com/eaburns/Watch)
* [https://github.com/alloy/kicker](https://github.com/alloy/kicker)
* [https://github.com/eradman/entr](https://github.com/eradman/entr)
* [https://github.com/cespare/reflex](https://github.com/cespare/reflex)
* [https://github.com/mlbileschi/file_system_monitor](https://github.com/mlbileschi/file_system_monitor)
