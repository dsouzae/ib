package ib

type InstrumentManager struct {
	AbstractManager
	id   int64
	c    Contract
	last float64
	bid  float64
	ask  float64
	size int64
}

func NewInstrumentManager(e *Engine, c Contract) (*InstrumentManager, error) {
	am, err := NewAbstractManager(e)
	if err != nil {
		return nil, err
	}

	m := &InstrumentManager{
		AbstractManager: *am,
		c:               c,
	}

	go m.startMainLoop(m.preLoop, m.receive, m.preDestroy)
	return m, nil
}

func (i *InstrumentManager) preLoop() error {
	i.id = i.eng.NextRequestId()
	req := &RequestMarketData{Contract: i.c}
	req.SetId(i.id)
	i.eng.Subscribe(i.rc, i.id)
	return i.eng.Send(req)
}

func (i *InstrumentManager) preDestroy() {
	i.eng.Unsubscribe(i.rc, i.id)
	req := &CancelMarketData{}
	req.SetId(i.id)
	i.eng.Send(req)
}

func (i *InstrumentManager) receive(r Reply) (UpdateStatus, error) {
	doupdate := false
	switch r.(type) {
	case *ErrorMessage:
		r := r.(*ErrorMessage)
		if r.SeverityWarning() {
			return UpdateFalse, nil
		}
		return UpdateFalse, r.Error()
	case *TickPrice:
		r := r.(*TickPrice)
		switch r.Type {
		case TickLast:
			doupdate = true
			i.last = r.Price
			i.size = r.Size
		case TickBid:
			doupdate = true
			i.bid = r.Price
			i.size = r.Size
		case TickAsk:
			doupdate = true
			i.ask = r.Price
			i.size = r.Size
		case TickLastTimestamp:
			doupdate = true
		case TickLastSize:
			doupdate = true
		}
	default:
	}

	if doupdate {
		return UpdateTrue, nil
	}

	return UpdateFalse, nil
}

func (i *InstrumentManager) Bid() float64 {
	i.rwm.RLock()
	defer i.rwm.RUnlock()
	return i.bid
}

func (i *InstrumentManager) Ask() float64 {
	i.rwm.RLock()
	defer i.rwm.RUnlock()
	return i.ask
}

func (i *InstrumentManager) Last() float64 {
	i.rwm.RLock()
	defer i.rwm.RUnlock()
	return i.last
}
func (i *InstrumentManager) Size() int64 {
	i.rwm.RLock()
	defer i.rwm.RUnlock()
	return i.size
}
