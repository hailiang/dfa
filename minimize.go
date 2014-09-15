package dfa

func (m *M) deleteUnreachable() *M {
	m.eachUnreachable(func(i int) {
		m.States.each(func(s *S) {
			a := s.Table.toTransArray()
			for b := range a {
				if a[b] == i {
					a[b] = invalidID //excludingID // TODO change this hack
				}
			}
			s.Table = a.toTransTable()
		})
	})
	return m
}
func (m *M) eachUnreachable(visit func(int)) {
	reachFinal := make([]bool, m.States.count())
	for i := range m.States {
		if m.States[i].final() {
			reachFinal[i] = true
		}
	}
	more := true
	for more {
		more = false
		for i := range reachFinal {
			if !reachFinal[i] {
				for j := range m.States[i].Table {
					next := m.States[i].Table[j].Next
					if next >= 0 && reachFinal[next] {
						reachFinal[i] = true
						more = true
						break
					}
				}
			}
		}
	}
	for i, r := range reachFinal {
		if !r {
			visit(i)
		}
	}
}

func (m *M) minimize() (*M, error) {
	if m == nil {
		return nil, nil
	}
	n := m.States.count()
	diff := newDiff(n)
	diff.eachFalse(func(i, j int) {
		s, t := m.States[i], m.States[j]
		if s.Label != t.Label || !s.Table.positionEqual(&t.Table) {
			diff.set(i, j)
		}
	})
	for diff.hasNewDiff {
		diff.hasNewDiff = false
		diff.eachFalse(func(i, j int) {
			s, t := m.States[i], m.States[j]
			si, ti := s.iter(), t.iter()
			_, sid := si.next()
			_, tid := ti.next()
			for sid != -1 && tid != -1 {
				if sid != tid && diff.get(sid, tid) {
					diff.set(i, j)
					break
				}
				_, sid = si.next()
				_, tid = ti.next()
			}
		})
	}
	idm := make(map[int]int)
	diff.eachFalse(func(i, j int) {
		idm[j] = i
	})
	if len(idm) > 0 {
		m.each(func(s *S) {
			s.each(func(t *Trans) {
				if small, ok := idm[t.Next]; ok {
					t.Next = small
				}
			})
		})
	}
	return m.or(m) // m.or(m) is also a way to remove unreachable nodes
}

func (t *TransTable) positionEqual(o *TransTable) bool {
	ti, oi := t.iter(), o.iter()
	for {
		tb, tnext := ti.next()
		ob, onext := oi.next()
		if tnext < 0 || onext < 0 {
			return tnext == onext
		}
		if tb != ob {
			return false
		}
	}
	return true
}

// 0: 1, 2, ..., n-1
// 1:    2, ..., n-1
// ...
// n-2:          n-1
type boolPairs struct {
	n          int
	a          []bool
	hasNewDiff bool
}

func newDiff(n int) *boolPairs {
	return &boolPairs{n, make([]bool, n*(n-1)/2), false}
}

func (d *boolPairs) set(i, j int) {
	d.hasNewDiff = true
	d.a[d.index(i, j)] = true
}

func (d *boolPairs) get(i, j int) bool {
	return d.a[d.index(i, j)]
}

func (d *boolPairs) index(i, j int) int {
	if i == j {
		panic("i should never be equal to j")
	} else if i > j {
		i, j = j, i
	}
	return (2*d.n-i-1)*i/2 + (j - i - 1)
}

func (d *boolPairs) eachFalse(visit func(int, int)) {
	for i := d.n - 2; i >= 0; i-- { // reverse order so the smaller comes later
		for j := i + 1; j <= d.n-1; j++ {
			if !d.get(i, j) {
				visit(i, j)
			}
		}
	}
}
