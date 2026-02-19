package application

type defaultColumnSpec struct {
	Name     string
	Color    string
	Position int
}

func defaultColumnSpecs() []defaultColumnSpec {
	return []defaultColumnSpec{
		{Name: "Todo", Color: "#60A5FA", Position: 1},
		{Name: "Doing", Color: "#F59E0B", Position: 2},
		{Name: "Done", Color: "#22C55E", Position: 3},
	}
}
