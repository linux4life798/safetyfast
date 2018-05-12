package safetyfast

type AtomicContext interface {
	Atomic(commiter func())
}
