package common

type Position struct {
	X      float64 // Signed. Coordinate X in millimeter.
	Y      float64 // Signed. Coordinate Y in millimeter.
	Height float64 // Signed. Height(Coordinate Z) in millimeter.
}

// Master uses an implementation of MobilityManager interface to simulate the mobility of nodes.
type MobilityManager interface {

	// Help message on how to set parameters
	ParametersHelp() string

	// Configure the mobility manager with a set of parameters.
	Configure(map[string]interface{}) error

	// Generate a new Position based on the configuration of this MobilityManager.
	GenerateNewNode() *Position

	// Set the Position slice that is managed by this MobilityManager. MobilityManager takes responsibility of modifying the slice.
	// Indices of Position are nodes' identities. Index 0 is intended to be nil all the time. nil at other indices means the node with the identity does not exist.
	Initialize(nodes []*Position)
}

// Master uses an implementation of September interface to make decisions upon packets. Interference, Packet Loss, etc. should be modeled in an implementation of this interface.
// Name September is from TV series Fringe http://fringe.wikia.com/wiki/September
// for his supernatural ability to observe everything and to manipulate (nearly) every object.
type September interface {

	// Help message on how to set parameters
	ParametersHelp() string

	// Configure the mobility manager with a set of parameters.
	Configure(map[string]interface{}) error

	// Set the Position slice that is used by September to decide various things, e.g., whether any pair of nodes can communicate.
	// Indices of Position are nodes' identities. Index 0 is intended to be nil all the time. nil at other indices means the node with the identity does not exist.
	Initialize(nodes []*Position)

	// Used when a unicast packet as large as size(in bytes) is sent from source(identity) to destination(identity).
	// Returns whether the packet should be delivered.
	// Any modification to models (interference, etc.) should be done within this function.
	SendUnicast(source int, destination int, size int) bool

	// Used when a broadcast packet as large as size(in bytes) is sent from source(identity).
	// Returns a slice of non-nil identities of nodes that should receive this packet. For effiency, the returned slice is a sub-slice of underlying.
	// underlying is a slice that garantees length large enough to hold all nodes. It does nothing more than providing a dedicated space for returned identities from this method. It's intended for reducing workload of GC. Thus, this method should modify elements in underlying and the returned slice should be a sub-slice of underlying. 
	// Any modification to models (interference, etc.) should be done within this function.
	SendBroadcast(source int, size int, underlying []int) []int
}
