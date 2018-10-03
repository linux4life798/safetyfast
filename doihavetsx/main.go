// This simple program prints the name of the cpu and if it supports
// Intel RTM and HLE.
// It exits with status code 0 only is RTM and HLE are both supported.
package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/intel-go/cpuid"
)

func main() {
	hasRTM := cpuid.HasExtendedFeature(cpuid.RTM)
	hasHLE := cpuid.HasExtendedFeature(cpuid.HLE)

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	printyesno := func(v bool) {
		if v {
			fmt.Fprintln(w, "Yes")
		} else {
			fmt.Fprintln(w, "No")
		}
	}

	fmt.Fprintf(w, "CPU Brand:\t%s\n", cpuid.ProcessorBrandString)
	fmt.Fprint(w, "RTM:\t")
	printyesno(hasRTM)
	fmt.Fprint(w, "HLE:\t")
	printyesno(hasHLE)
	w.Flush()

	if hasRTM && hasHLE {
		os.Exit(0)
	} else {
		os.Exit(1)
	}
}
