package godrop

type Clone struct {
	sesh *Session
	// 2 buffered channels to prevent dead locks as they depend on each other
	readHeader       chan int
	readContent      chan Header
	transferComplete chan int
}

// CloneDir clones a directory shared by a godrop peer
func (c *Clone) CloneDir(dir string) error {

	return ReadTarball(c.sesh.reader, dir)

}
