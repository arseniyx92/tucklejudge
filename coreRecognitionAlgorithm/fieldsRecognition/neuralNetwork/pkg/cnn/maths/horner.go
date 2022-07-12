package maths

// CoordsToHorner convert from coords to an index, using formula i = i1 + d1*(i2 + d2*(i3 + ...
func CoordsToHorner(coords, dims []int) int {
	horner := 0
	product := 1

	for i := 0; i < len(coords); i++ {
		horner += coords[i] * product
		product *= dims[i]
	}
	return horner
}

// HornerToCoords convert from index to coords, working back from formula i = i1 + d1*(i2 + d2*(i3 + ...
func HornerToCoords(hornerIndex int, dims []int) []int {
	coords := make([]int, len(dims))

	for i := 0; i < len(coords); i++ {
		coords[i] = hornerIndex % dims[i]
		hornerIndex = (hornerIndex - coords[i]) / dims[i]
	}

	return coords
}
