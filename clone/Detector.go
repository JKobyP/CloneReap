package clone

// Detector is the interface
type Detector interface {
	Detect(string) ([]ClonePair, error)
}

func NewDetector() Detector {
	return NewCcfx()
}
