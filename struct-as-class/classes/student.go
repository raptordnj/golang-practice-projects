package classes

import (
	"fmt"
)

type Person struct {
	Name string
	Age  int
}

func (p *Person) Greet() string {
	return fmt.Sprintf("Hi, I am %s and I am %d yars old.\b", p.Name, p.Age)
}

type Student struct {
	Person
	School string
}

func (s *Student) Study() string {
	return fmt.Sprintf("%s studing at %s\n", s.Name, s.School)
}
