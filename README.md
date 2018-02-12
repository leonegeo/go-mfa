# mfacache
Support for using cached MFA creds in Go

This package provides the ability to read the MFA-based credentials cached
by the AWS CLI tools and use them to create a new AWS SDK Session.
This allows one to, for example, take the "read-MFA-from-stdin" nastiness
out of your application.

To cache a set of credentials, run any AWS CLI command (or use the supplied app,
via `$ mfacache set`.)

You must be using the `AWS_PROFILE` environment variable for this to work.

Then, you can call `mfacache.NewSession` from within your (nonw now-noninteractive
app). This will create a Session using the cached creds.

-mpg
