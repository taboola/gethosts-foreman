gethosts-foreman
================

Bash auto completion of ssh hosts from foreman

 1. Define GOPATH environment variable (if not already defined)
 1. $ go get github.com/tsliwowicz/gethosts-foreman
 1. $ cp $GOPATH/bin/gethosts-foreman /usr/local/bin (or some other directory in your search path)
 1. $ vi $GOPATH/src/github.com/tsliwowicz/gethosts-foreman/foreman-completion.sh-template (replace the placeholders and save)
 1. $ cp $GOPATH/src/github.com/tsliwowicz/gethosts-foreman/foreman-completion.sh-template /etc/profile.d/foreman-completion.sh
 1. $ source /etc/profile.d/foreman-completion.sh (required only if you want the automcomplete to be active in the existing session)
 
 
