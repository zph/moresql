Deploying
=

On Server
==

* Download release or compile binary for platform
* Follow [setup guide for MoreSQL](/README/#quickstart)
* Set environmental variables
* Run moresql under process manager


On Heroku
==

* Follow [setup guide for MoreSQL](/README/#quickstart)

* Create repository for deployment

* Create Procfile

Sample:
```
worker: ./moresql -tail -checkpoint -error-reporting "rollbar"
```

* Set the ENV variables according to README [section](/README/#environmental-variables-used-in-moresql)

* Download latest stable release of moresql or build yourself for the linux amd64 platform using cross compilation

* Commit that binary to deploy project

* Add null buildpack for using a binary on heroku

```
heroku buildpacks:set -r REMOTE_NAME https://github.com/ryandotsmith/null-buildpack.git#72915d8b59f0f089931b4ed3b9c9b6f1750c331a
```

Note: we pin to specific version of buildpack so future upgrades aren't automatically applied.

* Deploy to heroku with a git push

