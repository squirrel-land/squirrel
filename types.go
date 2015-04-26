package squirrel

// MobilityManager controls locations and defines model of mobility of each
// nodes. Master uses an implementation of MobilityManager interface to
// simulate the mobility of nodes.
type MobilityManager interface {

	// ParametersHelp prints help message on how to set parameters
	ParametersHelp() string

	// Configure configures the mobility manager with a set of parameters.
	Configure(map[string]interface{}) error

	// Initialize sets the PositionManager.
	Initialize(positionManager PositionManager)
}

// September is used by Master to make decisions upon packets.
//
// Interference, Packet Loss, etc. should be modeled in an implementation of
// this interface.
//
// Name September is from TV series Fringe http://fringe.wikia.com/wiki/September
// for his supernatural ability to observe everything and to manipulate (nearly) every object.
type September interface {

	// ParametersHelp prints help message on how to set parameters
	ParametersHelp() string

	// Configure configures the mobility manager with a set of parameters.
	Configure(map[string]interface{}) error

	// Initialize sets the PositionManager.
	Initialize(positionManager PositionManager)

	// SendUnicast is used when a unicast packet as large as size(in bytes) is
	// sent from source(identity) to destination(identity).
	//
	// Returns whether the packet should be delivered.
	//
	// Any modification to models (interference, etc.) should be done within this
	// function.
	SendUnicast(source int, destination int, size int) bool

	// SendBroadcast is used when a broadcast packet as large as size(in bytes)
	// is sent from source(identity).
	//
	// Returns a slice of non-nil identities of nodes that should receive this
	// packet. For effiency, the returned slice is a sub-slice of underlying.
	//
	// underlying is a slice that garantees length large enough to hold all
	// nodes. It does nothing more than providing a dedicated space for returned
	// identities from this method. It's intended for reducing workload of GC.
	// Thus, this method should modify elements in underlying and the returned
	// slice should be a sub-slice of underlying.
	//
	// Any modification to models (interference, etc.) should be done within this
	// function.
	SendBroadcast(source int, size int, underlying []int) []int
}

type Position struct {
	X      float64 // Signed. Coordinate X in millimeter.
	Y      float64 // Signed. Coordinate Y in millimeter.
	Height float64 // Signed. Height(Coordinate Z) in millimeter.
}

type PositionManager interface {
	Capacity() int

	// Get returns a copy of Position at given index. Avoid this if possible. It
	// causes copying Position struct.
	Get(index int) Position

	// Distance calculates Euclidean distance between positions at index1 and
	// index2.
	Distance(index1, index2 int) float64

	// SetPosition sets position at index to be pos. It copies X, Y, and Height
	// values from inside pos into internal slice. pos is left intact and safe to
	// be changed afterwards.
	SetPosition(index int, pos *Position)
	Set(index int, x, y, height float64)

	// Enable marks a node as enabled.
	Enable(index int)

	// Disable marks a node as disabled.
	Disable(index int)

	// IsEnabled returns whether a node is enabled.
	IsEnabled(index int) bool

	Enabled() []int

	// RegisterEnabledChanged registers a channel, which when a node is enabled
	// or disabled, is used to send a slice of indices of all enabled nodes.
	RegisterEnabledChanged(channel chan<- []int)
}
