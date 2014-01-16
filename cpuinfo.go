// Contains a helper function for getting named properties out of the /dev/cpuinfo file.
// The file is only opened once. Properties are stored per processor. Processor 0 should
// be guaranteed to be present.

package hwio

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// maps processor:property to value.
var cpuInfo map[string]string

// Look up property for a CPU. CPU's start at 0.
func CpuInfo(cpu int, property string) string {
	if cpuInfo == nil {
		loadCpuInfo()
	}

	return cpuInfo[fmt.Sprintf("%d:%s", cpu, property)]
}

func loadCpuInfo() {
	cpuInfo = make(map[string]string)

	file, e := os.Open("/proc/cpuinfo")
	if e != nil {
		return
	}

	currentCpu := ""

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// split on the first colon, and trim both sides
		i := strings.Index(line, ":")
		if i >= 0 {
			name := strings.Trim(line[0:i], " \t")
			value := strings.Trim(line[i+1:], " \t")

			if name == "processor" {
				currentCpu = value
			}
			cpuInfo[currentCpu+":"+name] = value
		}

	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}
}
