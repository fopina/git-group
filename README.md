# git-group
Easily clone all the repositories from a group or organization


## WIP

* [ ] basic PoC
  * `clone X Y` creates Y directory, saves clone options in Y/.git-group and clones every repo
  * `pull` inside a git-group directory (or subdirectory) will pull every repository already cloned - flag for cloning new ones
* [ ] gitlab integration
* [ ] github integration
* [ ] Support both login and providing access token directly (required for 2FA users)
* [ ] search filters for `clone` command
