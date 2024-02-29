# module-logger

This tool simply writes log loading entries into `/var/log/module.log`.

## Install from Source

- Clone the repository
- Navigate into the repository
- `make`
- `make install`

## Install from Release

Grab the latest release from Github, unarchive it, and drop it somewhere useful
like `/usr/local/sbin/module-logger`.

Due to it writing to `/var/log/`, it must be owned by root and have the setuid
bit set to allow any user to run it:

```bash
chown root:root /path/to/module-logger
chmod 4755 /path/to/module-logger
```
