package cache

// Z is a sorted set member together with its score.
type Z struct {
	// Score is a weight, which orders member inside sorted set.
	Score float64

	// Member is a value, stored in sorted set. Unique inside its key: repeated
	// add of the same member updates its score instead of adding a second one.
	Member string
}

// ZRangeBy is a score range for sorted set lookups.
type ZRangeBy struct {
	// Min is an inclusive lower bound: a number, "-inf" or "(5" for exclusive one.
	Min string

	// Max is an inclusive upper bound: a number, "+inf" or "(5" for exclusive one.
	Max string

	// Offset is a number of matched members to skip. Applied together with Count.
	Offset int64

	// Count limits, how many members are returned. Zero means "all of them".
	Count int64
}
