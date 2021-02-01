Log[server]less
===============

This is some experimental tooling built upon compact ranges which allows the
maintenance and querying of an entirely on-disk log.

The idea is to make logging infrastructure a bit more *nix like.

A few example tools are provided:
 - `sequence` this assigns sequence numbers to new entries
 - `integrate` this integrates any as-yet un-integrated sequence numbers into the log state
 - `client` this provides log proof verification

This experiment currently supports a file-system based storage for the log data,
and the client's "transport" is to simple read the data from the same storage
location.

TODO:
 [ ] Tests, *cough*
 [ ] Add simple HTTP server which serves exactly the same structure as the filesystem storage
 [ ] Update client to be able to read tree data from the filesystem via HTTP