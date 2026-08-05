package categories

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) List() []Category {
	return []Category{
		{Name: "Велосипеды", Group: "bikes", Characteristics: []string{"бренд", "модель", "ростовка", "рост райдера", "размер колес", "материал рамы"}},
		{Name: "Рамы", Group: "frames", Characteristics: []string{"бренд", "ростовка", "материал", "год"}},
		{Name: "Колёса", Group: "wheels", Characteristics: []string{"размер", "втулки", "обод", "тип покрышки"}},
		{Name: "Рули", Group: "parts", Characteristics: []string{"ширина", "материал", "тип"}},
		{Name: "Шатуны", Group: "parts", Characteristics: []string{"длина", "BCD", "звезда"}},
		{Name: "Сёдла", Group: "parts", Characteristics: []string{"ширина", "материал"}},
		{Name: "Покрышки", Group: "parts", Characteristics: []string{"размер", "ширина", "компаунд"}},
		{Name: "Педали", Group: "parts", Characteristics: []string{"тип", "материал"}},
		{Name: "Аксессуары", Group: "accessories", Characteristics: []string{"тип", "состояние"}},
		{Name: "Fixed Gear", Group: "bikes", Characteristics: []string{"ростовка", "передача", "flip-flop", "тормоза"}},
		{Name: "MTB", Group: "bikes", Characteristics: []string{"размер колес", "подвеска", "ход вилки"}},
		{Name: "BMX", Group: "bikes", Characteristics: []string{"размер рамы", "размер колес"}},
		{Name: "Road", Group: "bikes", Characteristics: []string{"ростовка", "групсет", "материал рамы"}},
		{Name: "Gravel", Group: "bikes", Characteristics: []string{"ростовка", "клиренс покрышек", "материал рамы"}},
		{Name: "Детские велосипеды", Group: "bikes", Characteristics: []string{"возраст", "размер колес"}},
	}
}
