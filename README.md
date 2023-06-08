Tupi-auth-key is a plugin for Tupi to allow the use of an ``Authorization``
header to authenticate a request. Autorization keys are stored hashed with
sha512 in a sqlite database.

Install
=======

To install tupi-auth-key first clone the code:

```sh
$ git clone https://github.com/jucacrispim/tupi-auth-key
```

And then build the code:

```sh
$ cd tupi-auth-key
$ make build
```

This will create two binaries: ``./build/auth_plugin.so`` and ``./build/tupi-auth-key``.
The first file is the plugin that will be used by Tupi. The second file is a small
command line tool to manage your keys.

Usage
=====

To use the plugin with tupi, in  your config file put:

```toml
...
AuthPlugin = "/path/to/auth_plugin.so"
AuthPluginConf = {
    "uri": "/path/to/db.sqlite?the=option&other=value"
}
...
```

To manage your keys use the command line tool:

```sh
$ ./build/tupi-auth-key -h
```
