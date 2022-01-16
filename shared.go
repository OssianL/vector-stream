package main

const windowWidth = 600
const windowHeight = 500

const (
	sizeOfVec2 = 8
	sizeOfRect = 16
)

type Rect struct {
	position Vec2
	size     Vec2
}

func (r *Rect) GetCorners() [4]Vec2 {
	return [4]Vec2{
		r.position,
		Vec2{r.position.X + r.size.X, r.position.Y},
		Vec2{r.position.X + r.size.X, r.position.Y + r.size.Y},
		Vec2{r.position.X, r.position.Y + r.size.Y},
	}
}

type Vec2 struct {
	X float64
	Y float64
	// W float64
}

func (v Vec2) Add(o Vec2) Vec2 {
	return Vec2{
		X: v.X + o.X,
		Y: v.Y + o.Y,
	}
}

func (v Vec2) MultiplyFloat(f float64) Vec2 {
	return Vec2{
		X: v.X * f,
		Y: v.Y * f,
	}
}

func (v Vec2) DivideFloat(f float64) Vec2 {
	return Vec2{
		X: v.X / f,
		Y: v.Y / f,
	}
}
