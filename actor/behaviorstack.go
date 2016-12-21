package actor

type behaviorStack []Receive

func (b *behaviorStack) Clear() {
	*b = (*b)[:0]
}

func (b *behaviorStack) Peek() (v Receive, ok bool) {
	l := b.Len()
	if l > 0 {
		ok = true
		v = (*b)[l-1]
	}
	return
}

func (b *behaviorStack) Push(v Receive) {
	*b = append(*b, v)
}

func (b *behaviorStack) Pop() (v Receive, ok bool) {
	l := b.Len()
	if l > 0 {
		l--
		ok = true
		v = (*b)[l]
		(*b)[l] = nil
		*b = (*b)[:l]
	}
	return
}

func (b *behaviorStack) Len() int {
	return len(*b)
}
