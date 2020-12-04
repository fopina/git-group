# git-group
Easily clone all the repositories from a group or organization

This utility requires `git` installed. It could use [go-git](https://github.com/go-git/go-git) but it would fatten the binary with very little benefit as `git` is expected to be installed.

## Usage

Download the latest release of `git-group` binary and install it your `PATH`.

You can use it with `git-group ...` or as git subcommand `git group` (no difference)

```
$ git group clone https://gitlab.com/someorg/somegroup
Enter Username: fopina
Enter Password:
[1 / 856] Cloning proj1
Cloning into 'proj1'...
remote: Enumerating objects: 29, done.
remote: Counting objects: 100% (29/29), done.
remote: Compressing objects: 100% (24/24), done.
remote: Total 29 (delta 4), reused 0 (delta 0), pack-reused 0
Receiving objects: 100% (29/29), done.
Resolving deltas: 100% (4/4), done.
[2 / 856] Cloning proj2
Cloning into 'proj2'...
remote: Enumerating objects: 10, done.
remote: Counting objects: 100% (10/10), done.
remote: Compressing objects: 100% (7/7), done.
remote: Total 10 (delta 2), reused 0 (delta 0), pack-reused 0
...
```
