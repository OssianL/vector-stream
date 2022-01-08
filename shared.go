package main

const windowWidth = 600
const windowHeight = 500

type Rect struct {
	position Vec2
	W        float64
	H        float64
}

func (r *Rect) GetCorners() [4]Vec2 {
	return [4]Vec2{
		r.position,
		Vec2{r.position.X + r.W, r.position.Y},
		Vec2{r.position.X + r.W, r.position.Y + r.H},
		Vec2{r.position.X, r.position.Y + r.H},
	}
}

type FixpointRect struct {
	X uint16
	Y uint16
	W uint16
	H uint16
}
