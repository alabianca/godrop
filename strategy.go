package godrop

type ConnectionStrategy interface {
	Connect(peer string) (*P2PConn, error)
}
