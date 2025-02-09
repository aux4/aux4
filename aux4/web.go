package aux4

import (
	"fmt"
	"os"
	"runtime"
)

func GetUserAgent() string {
  name := os.Getenv("AUX4_HOSTNAME") 
  if name == "" {
    name, _ = os.Hostname()
  }
	return fmt.Sprintf("aux4/%s (%s; %s; %s)", Version, runtime.GOOS, runtime.GOARCH, name)
}
