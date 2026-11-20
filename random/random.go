package random

import rand "math/rand/v2"

// Random implements Effect.Random.random :: Effect Number.
// The backend invokes it at the Effect execution point.
func Random() float64 {
	return rand.Float64()
}
