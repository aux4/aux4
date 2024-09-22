package aux4

import (
	"fmt"
	"os"
	"runtime"
)

func GetUserAgent() string {
  name, _ := os.Hostname()
	return fmt.Sprintf("aux4/%s (%s; %s; %s)", Version, runtime.GOOS, runtime.GOARCH, name)
}
