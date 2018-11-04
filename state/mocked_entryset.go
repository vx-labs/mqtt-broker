package state

var _ EntrySet = &MockedEntryList{}

func (e *MockedEntryList) Append(entry Entry) {
	mockedEntry := entry.(*MockedEntry)
	e.MockedEntries = append(e.MockedEntries, mockedEntry)
}
func (e *MockedEntryList) Length() int {
	return len(e.MockedEntries)
}
func (e *MockedEntryList) New() EntrySet {
	return &MockedEntryList{}
}
func (e *MockedEntryList) AtIndex(idx int) Entry {
	return e.MockedEntries[idx]
}
func (e *MockedEntryList) Set(idx int, entry Entry) {
	mockedEntry := entry.(*MockedEntry)
	e.MockedEntries[idx] = mockedEntry
}
func (e *MockedEntryList) Range(f func(idx int, entry Entry)) {
	for idx, entry := range e.MockedEntries {
		f(idx, entry)
	}
}
