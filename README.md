# tanuki
Self-hosted OPDS manga server

## FAQ

**Q: What features does it have?**

- OPDS API
- Multiple user accounts
- Support for:
  - `.zip` and `.cbz` archives
  - `.jpeg`, `.png`, `.webp`, `.tiff` and `.bmp` images
- Nested folders in library

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
Usage of tanuki:
  -config string
        Path to config.json file. Leave blank to use the default config

$ tanukictl -help
Usage: tanukictl [options] [command] args

Options:
  -host string
        Host address of tanuki (default "0.0.0.0")
  -port string
        Port tanuki's RPC handler is listening on (default "9001")

Commands:
  scan                                  Scan the library
  dump                                  Dump the store's state
  user add <name>                       Add a new user with the password provided via stdin
  user delete <name>                    Delete an existing user
  user edit name <old-name> <new-name>  Change a user's name
  user edit pass <name>                 Change a user's password provided via stdin

  $ tanukictl -port 5000 scan
  $ tanukictl -port 5000 dump
    // Scan the library, then dump the store's contents
    // We connect to a tanuki instance listening on a
    // standard host but a non-standard port (5000)

  $ tanukictl user edit name old-name new-name
  $ echo "new-password" | tanukictl user edit pass new-name
    // Edit a user's name, then their password
```

**Q: What does the config file look like?**

This is TOML-encoded.

```toml
host = '0.0.0.0'
http_port = 8001
rpc_port = 9001
data_path = './data'
library_path = './library'
scan_interval = '1h0m0s'
```

**Q: Where's my username and password?**

The default username and password are logged to `STDERR` 
on the initialisation of the store.

To change the password, run `echo "new-password" | tanukictl user edit pass name`. 
This works assuming that you are using the default configuration.

**Q: I'm not using the default config, how do I connect to the server?**

You can manually specify the host and port that the tanuki RPC
handler has bound to, via `-host` and `-port`.

**Q: Do you support standalone files?**

No, if you want to add an entry to the library it must exist
within its own folder.

**Q: Should I expose the RPC port?**

No! It is not protected by any authentication mechanisms.

**Q: Can I run this using Docker?**

Yes, look at the following `compose.yaml` file.

Notice that for this setup to work, you must create the config
file before the container is created, otherwise Docker attempts
to create it as a directory.

```yaml
---
services:
  tanuki:
    image: ghcr.io/fiwippi/tanuki:latest
    command: -config /config.toml
    ports:
      - "8001:8001"
    volumes:
      - ./config.toml:/config.toml
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
$ tanukictl scan
Scan complete in 2ms
```
