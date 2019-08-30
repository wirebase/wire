The code in this repository copies the minimal amount of code required to generate
a embeddable filesystem from the following projects:
- github.com/shurcooL/httpfs
- github.com/shurcooL/vfsgen

With the following reasons:
- the functionality didn't rectify being depedant on multiple other project which
  are not super active
- the provided functionality is so central to our user experience we would like
  to replace it by our own implementation at some point
- the vfsgen generate function writes to stderr without being able to disable it
- some of the packages appear to have tests but they just print without any
  assertions.
