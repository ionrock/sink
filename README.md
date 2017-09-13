# Sink #

Sink is a server that will listen for GitHub comments and perform
operations based on the content of the comment. Sink takes the
workflow
in [hootsuite/atlantis](https://github.com/hootsuite/atlantis) and
makes it generalized to support other systems that require similar
locking and pre-merge actions on GitHub PRs.
