package game

type Cell struct {
	Value int
	Fixed bool
}

type Board [9][9]Cell

func NewSampleBoard() Board {
	return Board{
		{
			{Value: 1, Fixed: true},
			{Value: 6, Fixed: false},
			{Value: 2, Fixed: false},
			{Value: 7, Fixed: false},
			{Value: 3, Fixed: true},
			{Value: 4, Fixed: true},
			{Value: 5, Fixed: false},
			{Value: 9, Fixed: false},
			{Value: 8, Fixed: true},
		},
		{
			{Value: 4, Fixed: false},
			{Value: 7, Fixed: true},
			{Value: 9, Fixed: false},
			{Value: 6, Fixed: true},
			{Value: 8, Fixed: true},
			{Value: 5, Fixed: false},
			{Value: 1, Fixed: false},
			{Value: 3, Fixed: true},
			{Value: 2, Fixed: false},
		},
		{
			{Value: 5, Fixed: false},
			{Value: 3, Fixed: false},
			{Value: 8, Fixed: true},
			{Value: 2, Fixed: true},
			{Value: 1, Fixed: true},
			{Value: 9, Fixed: false},
			{Value: 7, Fixed: true},
			{Value: 6, Fixed: false},
			{Value: 4, Fixed: true},
		},
		{
			{Value: 2, Fixed: false},
			{Value: 5, Fixed: true},
			{Value: 4, Fixed: true},
			{Value: 1, Fixed: false},
			{Value: 9, Fixed: true},
			{Value: 7, Fixed: false},
			{Value: 6, Fixed: true},
			{Value: 8, Fixed: true},
			{Value: 3, Fixed: false},
		},
		{
			{Value: 9, Fixed: true},
			{Value: 1, Fixed: true},
			{Value: 3, Fixed: false},
			{Value: 5, Fixed: true},
			{Value: 6, Fixed: false},
			{Value: 8, Fixed: true},
			{Value: 4, Fixed: false},
			{Value: 2, Fixed: true},
			{Value: 7, Fixed: false},
		},
		{
			{Value: 6, Fixed: false},
			{Value: 8, Fixed: true},
			{Value: 7, Fixed: false},
			{Value: 3, Fixed: true},
			{Value: 4, Fixed: false},
			{Value: 2, Fixed: false},
			{Value: 9, Fixed: false},
			{Value: 1, Fixed: false},
			{Value: 5, Fixed: true},
		},
		{
			{Value: 3, Fixed: true},
			{Value: 4, Fixed: false},
			{Value: 5, Fixed: true},
			{Value: 9, Fixed: true},
			{Value: 2, Fixed: false},
			{Value: 6, Fixed: true},
			{Value: 8, Fixed: true},
			{Value: 7, Fixed: true},
			{Value: 1, Fixed: true},
		},
		{
			{Value: 7, Fixed: false},
			{Value: 2, Fixed: false},
			{Value: 6, Fixed: true},
			{Value: 8, Fixed: false},
			{Value: 5, Fixed: false},
			{Value: 1, Fixed: false},
			{Value: 3, Fixed: false},
			{Value: 4, Fixed: true},
			{Value: 9, Fixed: false},
		},
		{
			{Value: 8, Fixed: false},
			{Value: 9, Fixed: false},
			{Value: 1, Fixed: true},
			{Value: 4, Fixed: false},
			{Value: 7, Fixed: true},
			{Value: 3, Fixed: false},
			{Value: 2, Fixed: true},
			{Value: 5, Fixed: false},
			{Value: 6, Fixed: false},
		},
	}
}
