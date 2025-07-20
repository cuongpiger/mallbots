package main

type Person struct {
	Name string
}

type Student struct {
	Person
	Grade int
}

func (s *Student) PrintInfo() string {
	return "Grade: " + string(s.Grade)
}

func (s *Person) PrintInfo() string {
	return "Name: " + s.Name
}

func main() {
	

	// Calls Student's PrintInfo method
	println(student.PrintInfo()) // Output: Grade: 10

	// Calls Person's PrintInfo method
	println(student.Person.PrintInfo()) // Output: Name: Alice
}