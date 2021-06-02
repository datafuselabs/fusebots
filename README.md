<div align="center">
<h1>FuseBots</h1>
<img src='https://avatars.githubusercontent.com/u/82190365?s=120&v=4'>
</div>


## What I Can

* Nightly Release
  - release.yml
  - [example](https://github.com/datafuselabs/datafuse/releases)
  
* Auto Label
  - labeler.yml
  - [example](https://github.com/datafuselabs/datafuse/pulls?q=is%3Apr+is%3Aopen+label%3Apr-feature)
  
* Auto Merge
  - ALL CI passed
  - one of Reviewers APPROVED
  - [example](https://github.com/datafuselabs/datafuse/pull/636#issuecomment-849408422)
  
* Assistant
  - `/assginme` -- assign the issue to the user, [example](https://github.com/datafuselabs/datafuse/issues/663#issuecomment-851260591)

## Take me
```
go build cmd/fusebots
./fusebots -c your-config.ini
```
Set up `[host]:3000` to your webhook on GitHub.
