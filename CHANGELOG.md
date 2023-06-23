# Changelog

Starting with v0.24.0 I'm putting in a simple changelog because I'm starting to
forget which things I changed and why, even in this tiny repo.

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
