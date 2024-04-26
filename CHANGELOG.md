# Changelog

Starting with v0.24.0 I'm putting in a simple changelog because I'm starting to
forget which things I changed and why, even in this tiny repo.

# v0.28.0

- New hasher.FromString method for easier hasher.Hasher creation
- fileutil/manifest API changed again - Build is now just a non-hashed
  manifest; BuildHashed is how you say you want a hashed manifest.
- manifest example now verifies if your algorithm is valid when doing a
  `create-*` operation

# v0.27.0

v0.26.0 was kind of a flop and immediately needed a fix, so don't use it.

- bagit now requires a hash algorithm to be passed into `New` rather than
  defaulting to SHA256.
- bagit's Hasher is now in its own package, `hasher`, allowing non-bagit things
  to make use of the same general-purpose file hashing.
- fileutil/manifest has been massively refactored. There are two new top-level
  methods, `Build` and `Open`; it supports optional file hashing for doing more
  careful validation of a directory; and it has a whole suite of unit tests
  finally.

# v0.25.0

bagit has been given a way to provide a cache to avoid recomputing expensive
file checksums. By default no caching is done, and the API didn't change; a new
field was simply added.

# v0.24.0

We've reverted the auto-retry behavior of `fileutil.SyncDirectory` added in
v0.23.0. This can cause excessive wait times when network is unstable, which is
a massive pain: an error is often better than a process that should take a few
minutes ending up taking hours before *still* failing.

The bigger problem is that we've observed file loss with *no errors reported*.
We don't know if this is related to the silent retry, but it's one of a very
small number of changes which happened. My working theory (which seems
implausible, but at this point everything seems implausible) is that the
network has some kind of failure, but the file write operations end up retrying
enough to go to some phantom location that persists long enough to be read
(such as an in-memory cache that backs the filesystem mount), but never ends up
properly flushing to disk.

Either way, removing this feature because even if the second problem is
unrelated, the first problem is obnoxious.
