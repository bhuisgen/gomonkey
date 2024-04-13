# Build

This document describes how to build the SpiderMonkey library and how to include it to GoMonkey project.

## Prerequisites

Install Clang compiler and the required tooling:

```bash
$ sudo apt install clang clang-format make m4
```

Install Rust:

```bash
$ curl https://sh.rustup.rs -sSf | sh
$ source $HOME/.cargo/env
$ rustc --version
```

Install build dependencies:

```bash
$ sudo apt install libc++-dev libffi-dev zlib1g-dev
```

Install debug dependencies:

```bash
$ sudo apt install llvm-16 valgrind
$ sudo ln -s /usr/bin/llvm-objdump-16 /usr/local/bin/llvm-objdump
```

## Build

Download and extract the latest [source](https://ftp.mozilla.org/pub/firefox/releases/115.9.1esr/source/) of Mozilla Firefox ESR in your working directory.

```bash
$ tar xf firefox-115.9.1.tar.xz
$ cd firefox-115.9.1
```

First apply the patches present in the **patch** directory depending of the required OS and architecture:

```bash
$ cp ~/gomonkey/docs/patch/base/* .
$ for i in *.diff; do patch -p0 < $i; done
```

Only the build of the JS library is required:

```bash
$ cd ~/js/src
```

Create the build output directory:

```bash
$ mkdir _build
$ cd _build
```

**Debug**

To build the library with debugging symbols and sanity checks:

```bash
$ ../configure --disable-jemalloc --with-system-zlib --with-intl-api --enable-debug --disable-optimize --enable-hardening --enable-gc-probes --enable-gczeal --enable-valgrind --prefix=/workspace/gomonkey/deps/lib/linux_amd64/debug
$ make
$ make install
```

**Release**

To build the library for release:

```bash
$ ../configure --disable-jemalloc --with-system-zlib --with-intl-api --disable-debug --enable-optimize --enable-hardening --enable-strip --prefix=/workspace/gomonkey/deps/lib/linux_amd64/release
$ make
$ make install
```

## Integration

Update the CGO directives at the top of the source file `gomonkey.go`:

* The library path must match to your library file.
* Some build options are required for the debug library.
