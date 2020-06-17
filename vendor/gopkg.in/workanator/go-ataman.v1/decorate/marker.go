package decorate

// Marker contains rune sequence which identifies part of format tags.
type Marker string

// NewMarker constructs new instance from the marker given. If the marker is
// empty the function will panic.
func NewMarker(marker string) Marker {
	if len(marker) == 0 {
		panic("marker is empty")
	}

	return Marker(marker)
}

func (mrkr Marker) String() string {
	return string(mrkr)
}
