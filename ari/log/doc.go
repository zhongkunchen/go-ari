/*Package log wrap the package `logrus` for easy using

Usage example:

  package main

  import (
   "boot/log"
  )

  func main() {
    logger := log.NewLoggerFactory().GetLogger(nil)
    logger.Debug("hello")
    logger.WithFields(logrus.Fields{"[sn]": "sn-1998"}).Warn("go")
  }

Output:
WARN[2016-09-28T09:01:16+08:00] go                                            [sn]=sn-1998
*/
package log
