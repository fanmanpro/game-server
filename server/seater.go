package server

type Seater interface {
	SetSize(size int)
	SetSides(count int)
	SetSeat(seat Seat)
}
