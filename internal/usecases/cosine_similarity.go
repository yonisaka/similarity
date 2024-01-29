package usecases

import (
	"errors"
	"math"
)

// dotProduct Function to calculate dot product
func dotProduct(vectorA, vectorB []float64) float64 {
	var product float64
	for i := range vectorA {
		product += vectorA[i] * vectorB[i]
	}
	return product
}

// magnitude Function to calculate magnitude of a vector
func magnitude(vector []float64) float64 {
	var sum float64
	for _, val := range vector {
		sum += val * val
	}
	return math.Sqrt(sum)
}

// cosineSimilarity Function to calculate cosine similarity
func cosineSimilarity(vectorA, vectorB []float64) (float64, error) {
	if len(vectorA) != len(vectorB) {
		return 0, errors.New("vectors must be of the same length")
	}

	magA := magnitude(vectorA)
	magB := magnitude(vectorB)

	if magA == 0 || magB == 0 {
		return 0, errors.New("one or both vectors are of zero magnitude")
	}

	dotProd := dotProduct(vectorA, vectorB)
	return dotProd / (magA * magB), nil
}
