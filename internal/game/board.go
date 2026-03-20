package game

type Cell struct {
	Value int
	Fixed bool
}

type Board [9][9]Cell

func NewSampleBoard() Board {
	values := [9][9]int{
		{1, 0, 0, 0, 3, 4, 0, 0, 8},
		{0, 7, 0, 6, 8, 0, 0, 3, 0},
		{0, 0, 8, 2, 1, 0, 7, 0, 4},
		{0, 5, 4, 0, 9, 0, 6, 8, 0},
		{9, 1, 0, 5, 0, 8, 0, 2, 0},
		{0, 8, 0, 3, 0, 0, 0, 0, 5},
		{3, 0, 5, 9, 0, 6, 8, 7, 1},
		{0, 0, 6, 0, 0, 0, 0, 4, 0},
		{0, 0, 1, 0, 7, 0, 2, 0, 0},
	}

	var board Board
	for row := range 9 {
		for col := range 9 {
			board[row][col] = Cell{
				Value: values[row][col], 
				Fixed: values[row][col] != 0,
			}
		}
	}
	return board
}