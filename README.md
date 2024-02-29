# module-logger

This tool simply writes log loading entries into `/var/log/module.log`.

It must be owned by root and have the setuid bit set to allow any user to run it:

```bash
chown root:root /path/to/module-logger
chmod 4755 /path/to/module-logger
```

## Installation

Simply grab the latest release from Github.
