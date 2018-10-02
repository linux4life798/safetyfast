package safetyfast

func ExampleHLE() {

	m := map[string]int{
		"word1": 0,
		"word2": 0,
		"word3": 0,
		"word4": 0,
	}

	var lock SpinHLEMutex
	lock.Lock()
	// Action to be done transactionally
	count := m["word1"]
	m["word1"] = count + 1
	lock.Unlock()

}

func ExampleHLE2() {

	m := map[string]int{
		"word1": 0,
		"word2": 0,
		"word3": 0,
		"word4": 0,
	}

	c := NewLockedContext(new(SpinHLEMutex))
	c.Atomic(func() {
		// Action to be done transactionally
		count := m["word1"]
		m["word1"] = count + 1
	})

}

func ExampleRTM() {

	m := map[string]int{
		"word1": 0,
		"word2": 0,
		"word3": 0,
		"word4": 0,
	}

	c := NewRTMContexDefault()
	c.Atomic(func() {
		// Action to be done transactionally
		count := m["word1"]
		m["word1"] = count + 1
	})
}
