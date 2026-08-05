package categories

type Category struct {
	Name            string   `json:"name"`
	Group           string   `json:"group"`
	Characteristics []string `json:"characteristics"`
}
