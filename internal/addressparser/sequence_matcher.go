package addressparser

type matchBlock struct {
	a, b, size int
}

type sequenceMatcher struct {
	a, b []rune
	b2j  map[rune][]int
}

func newSequenceMatcher(a, b string) *sequenceMatcher {
	ra, rb := []rune(a), []rune(b)
	b2j := make(map[rune][]int)
	for j, r := range rb {
		b2j[r] = append(b2j[r], j)
	}

	return &sequenceMatcher{a: ra, b: rb, b2j: b2j}
}

func (sm *sequenceMatcher) findLongestMatch(alo, ahi, blo, bhi int) (int, int, int) {
	besti, bestj, bestsize := alo, blo, 0
	j2len := make(map[int]int)

	for i := alo; i < ahi; i++ {
		newj2len := make(map[int]int)
		for _, j := range sm.b2j[sm.a[i]] {
			if j < blo {
				continue
			}
			if j >= bhi {
				break
			}

			k := 1
			if prev, ok := j2len[j-1]; ok {
				k = prev + 1
			}
			newj2len[j] = k

			if k > bestsize {
				besti = i - k + 1
				bestj = j - k + 1
				bestsize = k
			}
		}
		j2len = newj2len
	}

	return besti, bestj, bestsize
}

func (sm *sequenceMatcher) findMatchingBlocks(alo, ahi, blo, bhi int) []matchBlock {
	besti, bestj, bestsize := sm.findLongestMatch(alo, ahi, blo, bhi)

	if bestsize == 0 {
		return nil
	}

	var blocks []matchBlock
	blocks = append(blocks, sm.findMatchingBlocks(alo, besti, blo, bestj)...)
	blocks = append(blocks, matchBlock{a: besti, b: bestj, size: bestsize})
	blocks = append(blocks, sm.findMatchingBlocks(besti+bestsize, ahi, bestj+bestsize, bhi)...)

	return blocks
}

func (sm *sequenceMatcher) ratio() float64 {
	blocks := sm.findMatchingBlocks(0, len(sm.a), 0, len(sm.b))

	var matches int
	for _, block := range blocks {
		matches += block.size
	}

	total := len(sm.a) + len(sm.b)
	if total == 0 {
		return 1.0
	}

	return 2.0 * float64(matches) / float64(total)
}

func (sm *sequenceMatcher) longestMatchSize() int {
	_, _, size := sm.findLongestMatch(0, len(sm.a), 0, len(sm.b))
	return size
}
