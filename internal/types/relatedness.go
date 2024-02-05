package types

// StringAndRelatedness holds a text and its relatedness score.
type StringAndRelatedness struct {
	ID          uint
	QdrantID    string
	Text        string
	Relatedness float64
}
