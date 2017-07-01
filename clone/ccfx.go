package clone

import (
	"fmt"
	"log"
	"os/exec"
)

type Ccfx struct {
}

func NewCcfx() *Ccfx {
	return &Ccfx{}
}

func (c *Ccfx) Detect(root string) ([]ClonePair, error) {
	// Find all clones in the repository
	log.Printf("\nExecuting code clone detector\n")
	ccfx := exec.Command("ccfx", "D", "cpp", "-d", root)
	err := ccfx.Start()
	if err != nil {
		return nil, fmt.Errorf("%v while executing %v", err, ccfx)
	}
	err = ccfx.Wait()
	if err != nil {
		return nil, fmt.Errorf("%v while executing %v", err, ccfx)
	}
	out, err := exec.Command("ccfx", "P", "a.ccfxd").Output()
	if err != nil {
		return nil, err
	}

	// Parse the output
	return CloneParse(string(out))
}
