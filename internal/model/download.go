package model

type Download interface {
	Title() string
	Wait()
}
