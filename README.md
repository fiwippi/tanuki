# tanuki
Self-hosted OPDS manga server

## FAQ

**Q: What features does it have?**

- OPDS API
- Multiple user accounts
- Support for `.zip` and `.cbz`
- Nested folders in library
- Single binary (~18 MB)

**Q: What's the OPDS support like?**

- The route for the OPDS catalog is `/opds/v1.2/catalog`
- This is the current [OPDS](https://specs.opds.io/) 1.2 feature support:
    - [x] Basic Auth
    - [x] Catalog feed
    - [x] Downloading of archives
    - [x] Getting cover/thumbnail of entries
    - [x] Searching (via OpenSearch)
    - [x] Page streaming

**Q: Does it have a CLI?**

Yes.

```console
$ tanuki -help
Usage: tanuki [options] [command] args

Options:
  -config string
        Path to config.json file. Leave blank to use the default config

Commands:
  run                                   Run the server
  scan                                  Scan the library
  dump                                  Dump the store's state
  user add <name>                       Add a new user with the password provided via stdin
  user delete <name>                    Delete an existing user
  user edit name <old-name> <new-name>  Change a user's name
  user edit pass <name>                 Change a user's password provided via stdin

Examples:
  $ tanuki -config /path/to/config.json run
    // Run the server using a specific config

  $ tanuki scan
  $ tanuki dump
    // Scan the library, then dump the store's contents

  $ tanuki -config custom.json user edit name old-name new-name
  $ echo "new-password" | tanuki -config custom.json user edit pass new-name
    // Edit a user's name, then its password. Since the server we are 
    // connecting to exposes a custom RPC port, we also supply the 
    // config to the CLI, (which details the value of the port)
```

**Q: What does the config file look like?**

This is JSON-encoded.

```json
{
 "host": "0.0.0.0",
 "http_port": 8001,
 "rpc_port": 9001,
 "data_path": "./data",
 "library_path": "./library",
 "scan_interval": "1h0m0s"
}
```

**Q: Where's my username and password?**

The default username and password are logged to `STDERR` 
on the initialisation of the store.

To change the password, run `echo "new-password" | tanuki user edit pass name`. 
This works assuming that you are using the default configuration.

**Q: I'm not using the default config, how do I connect to the server?**

The config you feed into the server specifies the port that 
it's listening on to accept RPC requests, these are requests 
that the CLI makes to the server in order to edit its state. 
Make sure that if your server is running on a custom config,
then you still feed as the config as input to RPC-based commands
such as `user edit`.

**Q: Do you support standalone files?**

No, if you want to add an entry to the library it must exist
within its own folder.

**Q: Should I expose the RPC port?**

No! It is not protected by any authentication mechanisms.

**Q: Can I run this using Docker?**

Yes, look at the following `compose.yaml` file.

Notice that for this setup to work, you must place the config
file in the data folder before the container is created, i.e.
don't wait for Docker to automatically create the data folder 
for you.

```yaml
---
version: "3"

services:
  tanuki:
    image: ghcr.io/fiwippi/tanuki:latest
    command: -config /data/config.json run
    ports:
      - "8001:8001"
    volumes:
      - ./library/:/library
      - ./data:/data
```

**Q: How can I run RPC commands with Docker?**

First attach to the container.

```console
$ docker exec -it tanuki /bin/sh
```

Now you can run commands as you please.

```console
$ tanuki -config ./data/config.json scan
Scan complete in 2ms
```
