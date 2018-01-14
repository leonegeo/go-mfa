# mfacache
Cache provider for AWS MFA tokens in Go

This package provides the ability to create MFA-based credentials, store
them in a cache file, and then use them to create a new AWS SDK Session.
This allows one to, for example, take the "read-MFA-from-stdin" nastiness
out of your application.

To use, first call `mfacache.StoreCredentials`: this will ask you for your
MFA token and then store the credentials to `$HOME/.aws/mfacache-PROFILE.json`.
(You can use the supplied app to do this, via `$ mfacache set`.)

Then, you can call `mfacache.NewSession` from within your now-noninteractive
app. This will create a Session using the cached creds.

Note this is not an implementation of the infamous "file provider cache", as
described in https://github.com/aws/aws-sdk-go/issues/1329 and
https://github.com/wallix/awless/issues/109, although it admittedly did start
out that way.

-mpg
