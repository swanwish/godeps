# The deps management tool

This tool is for go projects, because the go project usually depends on many other packages, and we need to get the code to gopath before build the project.
This tool will get the packages to vendors folder.

## Usage

### Install godeps

```
go get github.com/swanwish/godeps
```

## Init godeps.json file

It's hard to add depends one by one, the init function can get the external packages from dev environment.
The function depends on $GOROOT and $GOPATH env.

```
godeps init
```

## Add depends

Add command support two parameters, the path is package path, the godeps tool will create this folder under vendor path.

```
godeps add -p=packagepath -o=originpath

# example
godeps add -p=github.com/urfave/cli -o=https://github.com/urfave/cli
```

## Delete depend

The delete command support delete a package from json file according to the path parameter

```
godeps delete -p=packagepath

# example
godeps delete -p=github.com/urfave/cli
```

## List current depends

```
godeps list
```

## Download and sync local packages

```
godeps sync
```