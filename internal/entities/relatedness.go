package entities

// StringAndRelatedness holds a text and its relatedness score.
type StringAndRelatedness struct {
	ID          uint
	Text        string
	Relatedness float64
}
