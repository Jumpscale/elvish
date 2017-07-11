# A friendly and expressive Unix shell

[![GoDoc](http://godoc.org/github.com/jumpscale/elvish?status.svg)](http://godoc.org/github.com/jumpscale/elvish)
[![Build Status on Travis](https://travis-ci.org/elves/elvish.svg?branch=master)](https://travis-ci.org/elves/elvish)
[![Coverage Status](https://coveralls.io/repos/github/elves/elvish/badge.svg?branch=master)](https://coveralls.io/github/elves/elvish?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/jumpscale/elvish)](https://goreportcard.com/report/github.com/jumpscale/elvish)
[![License](https://img.shields.io/badge/License-BSD%202--Clause-orange.svg)](https://opensource.org/licenses/BSD-2-Clause)
[![Twitter](https://img.shields.io/twitter/url/http/shields.io.svg?style=social)](https://twitter.com/RealElvishShell)

General discussions:
[![Gitter](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/elves/elvish-public)
[![Telegram Group](https://img.shields.io/badge/telegram%20group-join-blue.svg)](https://telegram.me/elvish)
[![#elvish on freenode](https://img.shields.io/badge/freenode-%23elvish-000000.svg)](https://webchat.freenode.net/?channels=elvish)

Development discussions:
[![Gitter](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/elves/elvish-dev)
[![Telegram Group](https://img.shields.io/badge/telegram%20group-join-blue.svg)](https://telegram.me/elvish_dev)
[![#elvish on freenode](https://img.shields.io/badge/freenode-%23elvish--dev-000000.svg)](https://webchat.freenode.net/?channels=elvish-dev)

Elvish is a cross-platform shell suitable for both interactive use and scripting. It features a full-fledged, non-POSIX-shell programming language with advanced features like namespacing and anonymous functions, and a powerful, fully programmable user interface that works well out of the box.

... which is not 100% true yet. Elvish is already suitable for most daily interactive use, but it is not yet complete. Contributions are more than welcome!

Oh and here is a logo, which happens to be how Elvish looks like when you type `elvish` into it:

[![logo](https://elvish.io/assets/logo.svg)](https://elvish.io/)

This README documents the development aspect of Elvish. Other information is to be found on the [website](https://elvish.io).


## Building Elvish

To build Elvish, you need

*   A Go toolchain >= 1.6.

*   Linux (with x86, x64 or amd64 CPU) or macOS (with reasonably new hardware).

    It's quite likely that Elvish works on BSDs and other POSIX operating systems, or other CPU architectures; this is not guaranteed due to the lack of good CI support and developers who use such OSes. Pull requests are welcome.

    Windows is **not** supported yet.

### The Correct Way

Elvish is a go-gettable package. To build Elvish, first set up your Go workspace according to [How To Write Go Code](http://golang.org/doc/code.html), and then run

```sh
go get github.com/jumpscale/elvish
```

### The Lazy Way

Here is something you can copy-paste into your terminal:

```sh
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin
mkdir -p $GOPATH

go get github.com/jumpscale/elvish

for f in ~/.bashrc ~/.zshrc; do
    printf 'export %s=%s\n' GOPATH '$HOME/go' PATH '$PATH:$GOPATH/bin' >> $f
done
```

The scripts sets up the Go workspace and runs `go get` for you. It assumes that you have a working Go installation and currently use `bash` or `zsh`.

### The Homebrew Way

Users of macOS can build Elvish using [Homebrew](http://brew.sh):

```sh
brew install --HEAD elvish
```

### Extending Elvish
There are two ways for extending Elvish
- Using .elv scripts:
    - Add any `.elv` script file in `eval` directory
    - Regenerate embeddedModules using gen-embedded-modules elvish script
    ```
    elvish -c "cd <YOUR_GOPATH>/src/github.com/jumpscale/elvish/eval;./gen-embedded-modules"
    ```
    - Rebuild Elvish, you will find all of your variables and methods exist without using namespaces

- Using Golang methods
    - Create new go package in `eval` directory (i.e jumpscale)
    - in jumpscale package add you new methods and register their namespaces
    ```go
    package jumpscale

    import (
        "github.com/jumpscale/elvish/eval"
    )
    
    func Namespace() eval.Namespace {
        ns := eval.Namespace{}
        eval.AddBuiltinFns(ns, fns...)
        return ns
    }
    
    var fns = []*eval.BuiltinFn{
        {"myfunc", myFunc},
    }
    
    // Simple method to print each argument in newline on the console
    func myFunc(ec *eval.EvalCtx, args []eval.Value, opts map[string]eval.Value) {
        out := ec.OutputChan()
        for _, arg := range args {
            out <- eval.String(arg.Repr(0))
        }
    }
    ```
    - import your new package in the `main.go` file
    ```go
    import ("github.com/arahmanhamdy/elvish/eval/jumpscale")
    ```
    - in `main.go` file modify extraModule variable to include your module
    ```
    extraModules := map[string]eval.Namespace{
      		.....
      		"jumpscale": jumpscale.Namespace(),
      		....
    }
    ```
    - Rebuild Elvish, you will be able to call your myfunc method using jumpscale namespace (i.e `jumpscale:myfunc`)

Note that you can still extend elvish without editing the source code by [importing modules](https://elvish.io/ref/language.html#importing-module-use)


## Name

In [roguelikes](https://en.wikipedia.org/wiki/Roguelike), items made by the elves have a reputation of high quality. These are usually called *elven* items, but I chose "elvish" because it ends with "sh", a long tradition of Unix shells. It also rhymes with [fish](https://fishshell.com), one of shells that influenced the philosophy of Elvish.

The word "Elvish" should be capitalized like a proper noun. However, when referring to the `elvish` command, use it in lower case with fixed-width font.

Whoever practices the Elvish way by either contributing to it or simply using it is called an **Elf**. (You might have guessed this from the name of the GitHub organization.) The official adjective for Elvish (as in "Pythonic" for Python, "Rubyesque" for Ruby) is **Elven**.
