package cmd

import "github.com/sirupsen/logrus"

// continueOrFatal stops startup when a required dependency could not be built.
func continueOrFatal(err error) {
	if err != nil {
		logrus.Fatal(err)
	}
}
