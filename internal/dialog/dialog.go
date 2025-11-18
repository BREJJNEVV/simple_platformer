package dialog

// package dialog

// Dialog представляет простую систему диалогов: набор строк с текущим индексом
type Dialog struct {
	Lines []string
	idx   int
	done  bool
}

// New создает новый диалог из набора строк
func New(lines ...string) *Dialog {
	return &Dialog{Lines: append([]string{}, lines...), idx: 0, done: len(lines) == 0}
}

// Current возвращает текущую строку или пустую строку если диалог завершен
func (d *Dialog) Current() string {
	if d == nil || d.done || d.idx < 0 || d.idx >= len(d.Lines) {
		return ""
	}
	return d.Lines[d.idx]
}

// Next продвигает диалог к следующей строке, если строк больше нет — помечает как завершенный
func (d *Dialog) Next() {
	if d == nil || d.done {
		return
	}
	d.idx++
	if d.idx >= len(d.Lines) {
		d.done = true
	}
}

// Reset сбрасывает диалог в начало
func (d *Dialog) Reset() {
	if d == nil {
		return
	}
	d.idx = 0
	d.done = len(d.Lines) == 0
}

// IsFinished возвращает true если диалог окончен
func (d *Dialog) IsFinished() bool {
	if d == nil {
		return true
	}
	return d.done
}
