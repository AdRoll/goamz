

TODO
====

here will go things that we are activly pursuing. 


Brain Storm
===========

This is an unsorted, list of things that are 'nice to do'. From this list we will move things up to the active TODO list. Please feel absolutely free to contribute.

1. SES: we need SES support. Either we can implement it from scratch or choose to incorporate one of the already existing libraries. options:
  * http://www.stathat.com/src/amzses

2. Investigate if we can install `localdynamodb` on the ci to run tests against it. This way we can run behavioral tests that test the actual behavior. And keep them up to date with updates of `localdynamodb`.

3. Graduate items in `/exp`

4. Discusse "two tiered api". Basically the idea is to have a 'low level' api that is one to one match of the AWS http api, and a `higher level` api that abstracts away the details and is as go idiomatic as possible hiding away hairy details.

5. Find contributors with experience with each submodule (aws service), and review and plan for new refactored and cleaned up api.

6. It will be very helpfull to have better docs, sample code, best practices (for some people, goamz is their first exposure to AWS, and having a few essential tips on credentials and IAM roles and the like is very helpful).

7. If you've given any talks, presentation on goamz please let us know. We'd love to feature your video, slides, website here.
