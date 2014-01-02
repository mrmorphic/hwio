// Contains a helper function for getting named properties out of the /dev/cpuinfo file.
// The file is only opened once. Duplicate properties are overridden, so on multi-processor
// systems the values will generally be those for the last processor.

package hwio

import (
	"bufio"
	"os"
	"strings"
)

var cpuInfo map[string]string

func CpuInfo(property string) string {
	if cpuInfo == nil {
		loadCpuInfo()
	}
	return cpuInfo[property]
}

func loadCpuInfo() {
	cpuInfo = make(map[string]string)

	file, e := os.Open("/proc/cpuinfo")
	if e != nil {
		return
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// split on the first colon, and trim both sides
		i := strings.Index(line, ":")
		if i >= 0 {
			name := strings.Trim(line[0:i], " \t")
			value := strings.Trim(line[i+1:], " \t")
			cpuInfo[name] = value
		}
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}
}
