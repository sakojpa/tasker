package database

// Task represents a single task entity with fields for identification, scheduling, description, and repetition rules.
type Task struct {
	ID      string `yaml:"id" json:"id" db:"id"`
	Date    string `yaml:"date" json:"date" db:"date"`
	Title   string `yaml:"title" json:"title" db:"title"`
	Comment string `yaml:"comment" json:"comment" db:"comment"`
	Repeat  string `yaml:"repeat" json:"repeat" db:"repeat"`
}
