package safetyfast

// AtomicContext is the interface provided by a synchronization primitive
// that is capable of running a functions in an atomic context.
type AtomicContext interface {
	// Atomic will execute commiter exactly once per call in a manor that
	// appears to be atomic with respect to other commiters launched from this
	// AtomicContext.
	Atomic(commiter func())
}
