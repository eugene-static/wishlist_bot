package entity

type User struct {
	ID       int64
	Name     string
	Password []byte
}
type Wish struct {
	ID      string
	Content string
	UserID  int64
}
